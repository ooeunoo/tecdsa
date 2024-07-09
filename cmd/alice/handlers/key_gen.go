package handlers

import (
	"fmt"
	"io"
	"log"
	"tecdsa/pkg/database/repository"
	"tecdsa/pkg/dkls/dkg"
	"tecdsa/pkg/network"
	"tecdsa/pkg/utils"
	pb "tecdsa/proto/keygen"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/ot/base/simplest"
	"github.com/coinbase/kryptology/pkg/zkp/schnorr"
	"github.com/pkg/errors"
)

type KeygenHandler struct {
	curve *curves.Curve
	repo  repository.SecretRepository
}

func NewKeygenHandler(repo repository.SecretRepository) *KeygenHandler {
	return &KeygenHandler{
		curve: curves.K256(),
		repo:  repo,
	}
}

func (h *KeygenHandler) HandleKeyGen(stream pb.KeygenService_KeyGenServer) error {
	alice := dkg.NewAlice(h.curve)

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch msg := in.Msg.(type) {
		case *pb.KeygenMessage_Round1Response:
			err = h.handleRound2(stream, alice, msg.Round1Response)
		case *pb.KeygenMessage_Round3Response:
			err = h.handleRound4(stream, alice, msg.Round3Response)
		case *pb.KeygenMessage_Round5Response:
			err = h.handleRound6(stream, alice, msg.Round5Response)
		case *pb.KeygenMessage_Round7Response:
			err = h.handleRound8(stream, alice, msg.Round7Response)
		case *pb.KeygenMessage_Round9Response:
			err = h.handleRound10(stream, alice, msg.Round9Response)
		default:
			err = fmt.Errorf("unexpected message type")
		}

		if err != nil {
			return err
		}
	}
}

func (h *KeygenHandler) handleRound2(stream pb.KeygenService_KeyGenServer, alice *dkg.Alice, msg *pb.Round1Response) error {
	fmt.Println("라운드2")
	round2Output, err := alice.Round2CommitToProof([32]byte(msg.Seed))
	if err != nil {
		log.Printf("Error in Round2CommitToProof: %v", err)
		return err
	}
	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round2Response{
			Round2Response: &pb.Round2Response{
				Seed:       round2Output.Seed[:],
				Commitment: round2Output.Commitment,
			},
		},
	})
}

func (h *KeygenHandler) handleRound4(stream pb.KeygenService_KeyGenServer, alice *dkg.Alice, msg *pb.Round3Response) error {
	fmt.Println("라운드4")
	schnorrProof, err := h.parseSchnorrProof(msg)
	if err != nil {
		return err
	}
	proof, err := alice.Round4VerifyAndReveal(schnorrProof)
	if err != nil {
		log.Printf("Error in Round4VerifyAndReveal: %v", err)
		return err
	}
	if err != nil {
		log.Printf("Error in Round4VerifyAndReveal: %v", err)
		return err
	}
	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round4Response{
			Round4Response: &pb.Round4Response{
				C:         proof.C.Bytes(),
				S:         proof.S.Bytes(),
				Statement: proof.Statement.ToAffineCompressed(),
			},
		},
	})
}

func (h *KeygenHandler) handleRound6(stream pb.KeygenService_KeyGenServer, alice *dkg.Alice, msg *pb.Round5Response) error {
	fmt.Println("라운드6")

	schnorrProof, err := h.parseSchnorrProof(msg)
	if err != nil {
		return err
	}
	compressedReceiversMaskedChoice, err := alice.Round6DkgRound2Ot(schnorrProof)
	if err != nil {
		return err
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round6Response{
			Round6Response: &pb.Round6Response{
				ReceiversMaskedChoices: compressedReceiversMaskedChoice,
			},
		},
	})
}

func (h *KeygenHandler) handleRound8(stream pb.KeygenService_KeyGenServer, alice *dkg.Alice, msg *pb.Round7Response) error {
	fmt.Println("라운드8")

	challenges := make([]simplest.OtChallenge, len(msg.OtChallenges))
	for i, c := range msg.OtChallenges {
		copy(challenges[i][:], c)
	}
	challengeResponse, err := alice.Round8DkgRound4Ot(challenges)
	if err != nil {
		return err
	}
	challengeResponseBytes := make([][]byte, len(challengeResponse))
	for i, cr := range challengeResponse {
		challengeResponseBytes[i] = cr[:]
	}
	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round8Response{
			Round8Response: &pb.Round8Response{
				OtChallengeResponses: challengeResponseBytes,
			},
		},
	})
}

func (h *KeygenHandler) handleRound10(stream pb.KeygenService_KeyGenServer, alice *dkg.Alice, msg *pb.Round9Response) error {
	fmt.Println("라운드10")

	challengeOpenings := make([]simplest.ChallengeOpening, len(msg.ChallengeOpenings))
	for i, co := range msg.ChallengeOpenings {
		// co는 []byte 타입이므로, 이를 [2][32]byte 타입으로 변환
		if len(co) != 2*32 {
			return fmt.Errorf("invalid challenge opening length")
		}
		copy(challengeOpenings[i][0][:], co[:32])
		copy(challengeOpenings[i][1][:], co[32:])
	}
	err := alice.Round10DkgRound6Ot(challengeOpenings)
	if err != nil {
		return err
	}

	aliceOutput := alice.Output()
	address, err := network.DeriveAddress(aliceOutput.PublicKey, network.Ethereum)
	if err != nil {
		return err
	}

	// ###################################
	// TODO: 보안적으로 안전한 데이터 저장 플로우 필요
	secretKey, _ := utils.GenerateSecretKey()
	if err := h.repo.StoreSecretShare(address, aliceOutput, secretKey); err != nil {
		return errors.Wrap(err, "failed to store secret share")
	}
	// ###################################

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round10Response{
			Round10Response: &pb.Round10Response{
				Success:   true,
				SecretKey: secretKey,
			},
		},
	})
}

func (h *KeygenHandler) parseSchnorrProof(msg interface{}) (*schnorr.Proof, error) {
	var c, S []byte
	var statement []byte

	switch m := msg.(type) {
	case *pb.Round3Response:
		c, S, statement = m.C, m.S, m.Statement
	case *pb.Round5Response:
		c, S, statement = m.C, m.S, m.Statement
	default:
		return nil, fmt.Errorf("unexpected message type for Schnorr proof")
	}

	scalar, err := h.curve.Scalar.SetBytes(c)
	if err != nil {
		return nil, err
	}
	s, err := h.curve.Scalar.SetBytes(S)
	if err != nil {
		return nil, err
	}
	stmt, err := h.curve.Point.FromAffineCompressed(statement)
	if err != nil {
		return nil, err
	}
	return &schnorr.Proof{
		C:         scalar,
		S:         s,
		Statement: stmt,
	}, nil
}
