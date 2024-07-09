package handlers

import (
	"fmt"
	"io"
	"log"
	"tecdsa/pkg/database/repository"
	deserializer "tecdsa/pkg/deserializers"
	"tecdsa/pkg/network"
	"tecdsa/pkg/utils"
	pb "tecdsa/proto/keygen"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/dkg"
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

	// deserialize
	round2Input, err := deserializer.DecodeDkgRound2Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 2")
	}

	// round task
	output, err := alice.Round2CommitToProof(round2Input)
	if err != nil {
		log.Printf("Error in Round2CommitToProof in Round 2")
		return err
	}

	round2Output, err := deserializer.EncodeDkgRound2Output(output)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 2")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round2Response{
			Round2Response: &pb.Round2Response{
				Payload: round2Output,
			},
		},
	})
}

func (h *KeygenHandler) handleRound4(stream pb.KeygenService_KeyGenServer, alice *dkg.Alice, msg *pb.Round3Response) error {
	fmt.Println("라운드4")
	round4Input, err := deserializer.DecodeDkgRound4Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 4")
	}

	// round task
	proof, err := alice.Round4VerifyAndReveal(round4Input)
	if err != nil {
		log.Printf("Error in Round4VerifyAndReveal in Round 4")
		return err
	}

	round4Output, err := deserializer.EncodeDkgRound4Output(proof)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 4")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round4Response{
			Round4Response: &pb.Round4Response{
				Payload: round4Output,
			},
		},
	})
}

func (h *KeygenHandler) handleRound6(stream pb.KeygenService_KeyGenServer, alice *dkg.Alice, msg *pb.Round5Response) error {
	fmt.Println("라운드6")

	round6Input, err := deserializer.DecodeDkgRound6Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 6")
	}
	// round task
	output, err := alice.Round6DkgRound2Ot(round6Input)
	if err != nil {
		return err
	}

	round6Output, err := deserializer.EncodeDkgRound6Output(output)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 6")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round6Response{
			Round6Response: &pb.Round6Response{
				Payload: round6Output,
			},
		},
	})
}

func (h *KeygenHandler) handleRound8(stream pb.KeygenService_KeyGenServer, alice *dkg.Alice, msg *pb.Round7Response) error {
	fmt.Println("라운드8")

	round8Input, err := deserializer.DecodeDkgRound8Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 8")
	}

	// round task
	challengeResponse, err := alice.Round8DkgRound4Ot(round8Input)
	if err != nil {
		return err
	}

	round8Output, err := deserializer.EncodeDkgRound8Output(challengeResponse)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 8")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round8Response{
			Round8Response: &pb.Round8Response{
				Payload: round8Output,
			},
		},
	})
}

func (h *KeygenHandler) handleRound10(stream pb.KeygenService_KeyGenServer, alice *dkg.Alice, msg *pb.Round9Response) error {
	fmt.Println("라운드10")

	round10Input, err := deserializer.DecodeDkgRound10Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 10")
	}

	// round task
	roundErr := alice.Round10DkgRound6Ot(round10Input)
	if roundErr != nil {
		return roundErr
	}

	// save result
	aliceOutput := alice.Output()
	address, err := network.DeriveAddress(aliceOutput.PublicKey, network.Ethereum)
	if err != nil {
		return err
	}

	// ###################################
	// TODO: 보안적으로 안전한 데이터 저장 플로우 필요
	secretKey, _ := utils.GenerateSecretKey()
	share, err := deserializer.EncodeAliceDkgOutput(aliceOutput)
	if err != nil {
		return errors.Wrap(err, "failed to encode alice output")
	}

	if err := h.repo.StoreSecretShare(address, share, secretKey); err != nil {
		return errors.Wrap(err, "failed to store secret alice share")
	}

	// ###################################

	// serialize
	round10Output := pb.Round10Response{
		Success:   true,
		SecretKey: secretKey,
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_Round10Response{
			Round10Response: &round10Output,
		},
	})
}
