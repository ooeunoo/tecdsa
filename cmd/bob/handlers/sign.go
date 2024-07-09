// bob handler
package handlers

import (
	"hash"
	"io"
	"tecdsa/pkg/database/repository"

	// "tecdsa/pkg/dkls/dkg"

	// "tecdsa/pkg/dkls/dkg"
	pb "tecdsa/proto/sign"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/sign"
	"golang.org/x/crypto/sha3"
)

type SignHandler struct {
	curve *curves.Curve
	hash  hash.Hash
	repo  repository.SecretRepository
	bob   *sign.Bob
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
		// case *pb.SignMessage_Round1Response:
		// 	err = h.handleRound2(stream, msg.Round1Response)
		// case *pb.SignMessage_Round3Response:
		// 	err = h.handleRound4(stream, msg.Round3Response)
		// default:
		// 	err = fmt.Errorf("unexpected message type")
		// }

		// if err != nil {
		// 	return err
		// }
	}
}

// func (h *SignHandler) handleRound2(stream pb.SignService_SignServer, msg *pb.Round1Response) error {
// 	fmt.Println("라운드2")

// 	output, err := h.repo.GetSecretShare(msg.Address, msg.SecretKey)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get secret share")
// 	}

// 	bobOutput, ok := output.(*repository.BobOutputWrapper)
// 	if !ok {
// 		return errors.New("retrieved secret share is not a BobOutput")
// 	}

// 	h.bob = sign.NewBob(h.curve, h.hash, bobOutput.BobOutput)

// 	seed, err := deserializer.DecodeSignRound2Input(msg.Payload)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to decode in Round 2")
// 	}

// 	round2Output, err := h.bob.Round2Initialize(seed)
// 	if err != nil {
// 		return errors.Wrap(err, "failed in Round2Initialize")
// 	}

// 	encodeRound2Ouput, err := deserializer.EncodeSignRound2Output(round2Output)
// 	if err != nil {
// 		return errors.Wrap(err, "failed in encode in Round 2")
// 	}

// 	return stream.Send(&pb.SignMessage{
// 		Msg: &pb.SignMessage_Round2Response{
// 			payload: encodeRound2Ouput,
// 		},
// 	})
// }

// func (h *SignHandler) handleRound4(stream pb.SignService_SignServer, msg *pb.Round3Response) error {
// 	fmt.Println("라운드4")

// 	// deserialize
// 	round4Input, err := deserializer.DecodeSignRound4Input(msg.Payload)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to decode in Round 4")
// 	}

// 	if err = h.bob.Round4Final("message", round4Input); err != nil {
// 		return errors.Wrap(err, "failed in Round4Final")
// 	}

// 	return stream.Send(&pb.SignMessage{
// 		Msg: &pb.SignMessage_Round4Response{
// 			Round4Response: &round4Response},
// 	})
// }
