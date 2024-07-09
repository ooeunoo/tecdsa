// alice handler
package handlers

import (
	"hash"
	"io"
	"tecdsa/pkg/database/repository"
	pb "tecdsa/proto/sign"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/sign"
	"golang.org/x/crypto/sha3"
)

type SignHandler struct {
	curve *curves.Curve
	hash  hash.Hash
	repo  repository.SecretRepository
	alice *sign.Alice
}

func NewSignHandler(repo repository.SecretRepository) *SignHandler {
	return &SignHandler{
		curve: curves.K256(),
		hash:  sha3.NewLegacyKeccak256(),
		repo:  repo,
	}
}

func (h *SignHandler) HandleSign(stream pb.SignService_SignServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// switch msg := in.Msg.(type) {
		// case *pb.SignMessage_Round1Request:
		// 	err = h.handleRound1(stream, msg.Round1Request)
		// case *pb.SignMessage_Round2Response:
		// 	err = h.handleRound3(stream, msg.Round2Response)
		// default:
		// 	err = fmt.Errorf("unexpected message type")
		// }

		// if err != nil {
		// 	return err
		// }
	}
}

// func (h *SignHandler) handleRound1(stream pb.SignService_SignServer, msg *pb.Round1Request) error {
// 	fmt.Println("라운드1")

// 	output, err := h.repo.GetSecretShare(msg.Address, msg.SecretKey)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get secret share")
// 	}

// 	aliceOutput, ok := output.(*dkg.AliceOutput)
// 	if !ok {
// 		return errors.New("retrieved secret share is not an AliceOutput")
// 	}

// 	h.alice = sign.NewAlice(h.curve, h.hash, aliceOutput)

// 	seed, err := h.alice.Round1GenerateRandomSeed()
// 	if err != nil {
// 		return errors.Wrap(err, "failed to generate random seed in Round 1")
// 	}

// 	// serialize
// 	encodedPayload, err := deserializer.EncodeSignRound1Output(seed)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to generate random seed in Round 1")
// 	}
// 	round1Response := pb.Round1Response{
// 		Payload: encodedPayload,
// 	}

// 	return stream.Send(&pb.SignMessage{
// 		Msg: &pb.SignMessage_Round1Response{
// 			Round1Response: &round1Response,
// 		},
// 	})
// }

// func (h *SignHandler) handleRound3(stream pb.SignService_SignServer, msg *pb.Round2Response) error {
// 	fmt.Println("라운드3")

// 	// decode
// 	round2Output, err := deserializer.DecodeSignRound3Input(msg.Payload)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to decode in Round 3 input")
// 	}

// 	round3Output, err := h.alice.Round3Sign([32]byte("sadfsadfsa"), round2Output)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to sign in Round 3")
// 	}

// 	// serialize
// 	encodedPayload, err := deserializer.EncodeSignRound3Output(round3Output)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to encode in Round 3 output")
// 	}
// 	round3Response := pb.Round3Response{
// 		Payload: encodedPayload,
// 	}
// 	// 결과를 클라이언트에게 전송
// 	return stream.Send(&pb.SignMessage{
// 		Msg: &pb.SignMessage_Round3Response{
// 			Payload: round3Response,
// 		},
// 	})

// }
