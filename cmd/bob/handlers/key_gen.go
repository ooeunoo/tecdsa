// dkg_handler.go
package handlers

import (
	"fmt"
	"io"
	"tecdsa/pkg/database/repository"
	deserializer "tecdsa/pkg/deserializers"
	"tecdsa/pkg/network"
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
		case *pb.KeygenMessage_KeyGenRequestTo1Output:
			err = h.handleRound1(stream, bob, msg.KeyGenRequestTo1Output)
		case *pb.KeygenMessage_KeyGenRound2To3Output:
			err = h.handleRound3(stream, bob, msg.KeyGenRound2To3Output)
		case *pb.KeygenMessage_KeyGenRound4To5Output:
			err = h.handleRound5(stream, bob, msg.KeyGenRound4To5Output)
		case *pb.KeygenMessage_KeyGenRound6To7Output:
			err = h.handleRound7(stream, bob, msg.KeyGenRound6To7Output)
		case *pb.KeygenMessage_KeyGenRound8To9Output:
			err = h.handleRound9(stream, bob, msg.KeyGenRound8To9Output)
		case *pb.KeygenMessage_KeyGenRound10To11Output:
			err = h.handleRound11(stream, bob, msg.KeyGenRound10To11Output)
		default:
			err = fmt.Errorf("unexpected message type")
		}

		if err != nil {
			return err
		}
	}
}

func (h *KeygenHandler) handleRound1(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.KeyGenRequestTo1Output) error {
	fmt.Println("라운드1")
	seed, err := bob.Round1GenerateRandomSeed()
	if err != nil {
		return errors.Wrap(err, "failed to Round1GenerateRandomSeed in Round 1")
	}

	round1Output, err := deserializer.EncodeDkgRound1Output(seed)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 1")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound1To2Output{
			KeyGenRound1To2Output: &pb.KeyGenRound1To2Output{
				Payload: round1Output,
			},
		},
	})
}

func (h *KeygenHandler) handleRound3(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.KeyGenRound2To3Output) error {
	fmt.Println("라운드3")

	// deserialize
	round3Input, err := deserializer.DecodeDkgRound3Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 3")
	}

	// round task
	proof, err := bob.Round3SchnorrProve(round3Input)
	if err != nil {
		return errors.Wrap(err, "failed to Round3SchnorrProve in Round3")
	}

	round3Output, err := deserializer.EncodeDkgRound3Output(proof)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 3")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound3To4Output{
			KeyGenRound3To4Output: &pb.KeyGenRound3To4Output{
				Payload: round3Output,
			},
		},
	})
}

func (h *KeygenHandler) handleRound5(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.KeyGenRound4To5Output) error {
	fmt.Println("라운드5")

	round5Input, err := deserializer.DecodeDkgRound5Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 5")
	}

	proof, err := bob.Round5DecommitmentAndStartOt(round5Input)
	if err != nil {
		return errors.Wrap(err, "failed in Round5DecommitmentAndStartOt in Round 5")
	}

	round5Output, err := deserializer.EncodeDkgRound5Output(proof)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 5")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound5To6Output{
			KeyGenRound5To6Output: &pb.KeyGenRound5To6Output{
				Payload: round5Output,
			},
		},
	})
}

func (h *KeygenHandler) handleRound7(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.KeyGenRound6To7Output) error {
	fmt.Println("라운드7")

	round7Input, err := deserializer.DecodeDkgRound7Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 7")
	}

	// round task
	challenges, err := bob.Round7DkgRound3Ot(round7Input)
	if err != nil {
		return errors.Wrap(err, "failed in Round7DkgRound3Ot in Round 7")
	}

	round7Output, err := deserializer.EncodeDkgRound7Output(challenges)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 5")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound7To8Output{
			KeyGenRound7To8Output: &pb.KeyGenRound7To8Output{
				Payload: round7Output,
			},
		},
	})
}

func (h *KeygenHandler) handleRound9(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.KeyGenRound8To9Output) error {
	fmt.Println("라운드9")
	round9Input, err := deserializer.DecodeDkgRound9Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 9")
	}

	// round task
	challengeOpenings, err := bob.Round9DkgRound5Ot(round9Input)
	if err != nil {
		return errors.Wrap(err, "failed in Round9DkgRound5Ot in Round 9")
	}

	round9Output, err := deserializer.EncodeDkgRound9Output(challengeOpenings)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 5")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound9To10Output{
			KeyGenRound9To10Output: &pb.KeyGenRound9To10Output{
				Payload: round9Output,
			}},
	})
}

func (h *KeygenHandler) handleRound11(stream pb.KeygenService_KeyGenServer, bob *dkg.Bob, msg *pb.KeyGenRound10To11Output) error {
	fmt.Println("라운드끝")

	bobOutput := bob.Output()
	address, err := network.DeriveAddress(bobOutput.PublicKey, network.Ethereum)
	if err != nil {
		return err
	}

	// ###################################
	// TODO: 보안적으로 안전한 데이터 저장 플로우 필요

	secretKey := msg.SecretKey
	share, err := deserializer.EncodeBobDkgOutput(bobOutput)
	if err != nil {
		return errors.Wrap(err, "failed to encode bob output")
	}

	if err := h.repo.StoreSecretShare(address, share, secretKey); err != nil {
		return errors.Wrap(err, "failed to store secret bob share")
	}
	// ###################################

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound11ToResponseOutput{
			KeyGenRound11ToResponseOutput: &pb.KeyGenRound11ToResponseOutput{
				Address:   address,
				SecretKey: secretKey,
			},
		},
	})
}
