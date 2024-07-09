// dkg_handler.go
package handlers

import (
	"fmt"
	"io"
	"tecdsa/pkg/database/repository"
	"tecdsa/pkg/dkls/dkg"
	"tecdsa/pkg/network"
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
	bob := dkg.NewBob(h.curve)

	for {
		in, err := stream.Recv()

		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch msg := in.Msg.(type) {
		case *pb.KeygenMessage_Round1Request:
			err = h.handleRound1(stream, bob)
		case *pb.KeygenMessage_Round2Response:
			err = h.handleRound3(stream, bob, msg.Round2Response)
		case *pb.KeygenMessage_Round4Response:
			err = h.handleRound5(stream, bob, msg.Round4Response)
		case *pb.KeygenMessage_Round6Response:
			err = h.handleRound7(stream, bob, msg.Round6Response)
		case *pb.KeygenMessage_Round8Response:
			err = h.handleRound9(stream, bob, msg.Round8Response)
		case *pb.KeygenMessage_Round10Response:
			err = h.handleFinalRound(stream, bob, msg.Round10Response)
		default:
			err = fmt.Errorf("unexpected message type")
		}

		if err != nil {
			return err
		}
	}
}

func (h *KeygenHandler) handleRound1(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob) error {
	fmt.Println("라운드1")
	seed, err := bob.Round1GenerateRandomSeed()
	if err != nil {
		return errors.Wrap(err, "failed to generate random seed in Round 1")
	}
	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round1Response{
			Round1Response: &pb.Round1Response{Seed: seed[:]},
		},
	})
}

func (h *KeygenHandler) handleRound3(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.Round2Response) error {
	fmt.Println("라운드3")
	proof, err := bob.Round3SchnorrProve(&dkg.Round2Output{
		Seed:       [32]byte(msg.Seed),
		Commitment: msg.Commitment,
	})
	if err != nil {
		return errors.Wrap(err, "failed in Round3SchnorrProve")
	}
	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round3Response{
			Round3Response: &pb.Round3Response{
				C:         proof.C.Bytes(),
				S:         proof.S.Bytes(),
				Statement: proof.Statement.ToAffineCompressed(),
			},
		},
	})
}

func (h *KeygenHandler) handleRound5(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.Round4Response) error {
	fmt.Println("라운드5")
	schnorrProof, err := h.parseSchnorrProof(msg)
	if err != nil {
		return err
	}
	proof, err := bob.Round5DecommitmentAndStartOt(schnorrProof)
	if err != nil {
		return errors.Wrap(err, "failed in Round5DecommitmentAndStartOt")
	}
	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round5Response{
			Round5Response: &pb.Round5Response{
				C:         proof.C.Bytes(),
				S:         proof.S.Bytes(),
				Statement: proof.Statement.ToAffineCompressed(),
			}},
	})
}

func (h *KeygenHandler) handleRound7(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.Round6Response) error {
	fmt.Println("라운드7")
	compressedReceiversMaskedChoice := make([]simplest.ReceiversMaskedChoices, len(msg.ReceiversMaskedChoices))
	for i, choice := range msg.ReceiversMaskedChoices {
		compressedReceiversMaskedChoice[i] = choice
	}
	challenges, err := bob.Round7DkgRound3Ot(compressedReceiversMaskedChoice)
	if err != nil {
		return errors.Wrap(err, "failed in Round7DkgRound3Ot")
	}
	challengesBytes := make([][]byte, len(challenges))
	for i, challenge := range challenges {
		challengesBytes[i] = challenge[:]
	}
	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round7Response{
			Round7Response: &pb.Round7Response{
				OtChallenges: challengesBytes,
			},
		},
	})
}

func (h *KeygenHandler) handleRound9(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.Round8Response) error {
	fmt.Println("라운드9")
	challengeResponses := make([]simplest.OtChallengeResponse, len(msg.OtChallengeResponses))
	for i, response := range msg.OtChallengeResponses {
		copy(challengeResponses[i][:], response)
	}
	challengeOpenings, err := bob.Round9DkgRound5Ot(challengeResponses)
	if err != nil {
		return errors.Wrap(err, "failed in Round9DkgRound5Ot")
	}
	challengeOpeningsBytes := make([][]byte, len(challengeOpenings))
	for i, opening := range challengeOpenings {
		challengeOpeningsBytes[i] = make([]byte, 2*32)
		copy(challengeOpeningsBytes[i][0:32], opening[0][:])
		copy(challengeOpeningsBytes[i][32:], opening[1][:])
	}
	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round9Response{
			Round9Response: &pb.Round9Response{
				ChallengeOpenings: challengeOpeningsBytes,
			},
		},
	})
}

func (h *KeygenHandler) handleFinalRound(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.Round10Response) error {
	fmt.Println("라운드끝")
	bobOutput := bob.Output()
	address, err := network.DeriveAddress(bobOutput.PublicKey, network.Ethereum)
	if err != nil {
		return err
	}

	// ###################################
	// TODO: 보안적으로 안전한 데이터 저장 플로우 필요
	secretKey := msg.SecretKey
	if err := h.repo.StoreSecretShare(address, bobOutput, secretKey); err != nil {
		return errors.Wrap(err, "failed to store secret share")
	}
	// ###################################

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenResponse{
			KeyGenResponse: &pb.KeyGenResponse{
				Success:   true,
				Address:   address,
				SecretKey: secretKey,
			},
		},
	})
}

func (h *KeygenHandler) parseSchnorrProof(msg *pb.Round4Response) (*schnorr.Proof, error) {
	c, err := h.curve.Scalar.SetBytes(msg.C)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set C bytes")
	}
	S, err := h.curve.Scalar.SetBytes(msg.S)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set S bytes")
	}
	statement, err := h.curve.Point.FromAffineCompressed(msg.Statement)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set Statement bytes")
	}
	return &schnorr.Proof{
		C:         c,
		S:         S,
		Statement: statement,
	}, nil
}
