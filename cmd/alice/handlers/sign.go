// dkg_handler.go
package handlers

import (
	"fmt"
	"hash"
	"io"
	"tecdsa/pkg/database/repository"
	"tecdsa/pkg/dkls/dkg"
	"tecdsa/pkg/dkls/sign"
	pb "tecdsa/proto/sign"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/pkg/errors"
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
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch msg := in.Msg.(type) {
		case *pb.SignMessage_Round1Request:
			err = h.handleRound1(stream, msg.Round1Request)
		case *pb.SignMessage_Round2Response:
			// err = h.handleRound3(stream, alice, msg.Round2Response)
		default:
			err = fmt.Errorf("unexpected message type")
		}

		if err != nil {
			return err
		}
	}
}

func (h *SignHandler) handleRound1(stream pb.SignService_SignServer, msg *pb.Round1Request) error {
	fmt.Println("라운드1")
	// GetSecretShare 함수를 사용하여 AliceOutput 가져오기
	aliceOutputInterface, err := h.repo.GetSecretShare(msg.Address, msg.SecretKey)
	if err != nil {
		return errors.Wrap(err, "failed to get secret share")
	}

	aliceOutput, ok := aliceOutputInterface.(*dkg.AliceOutput)
	if !ok {
		return errors.New("retrieved secret share is not an AliceOutput")
	}

	// Alice 인스턴스 생성
	h.alice = sign.NewAlice(h.curve, h.hash, aliceOutput)

	seed, err := h.alice.Round1GenerateRandomSeed()
	if err != nil {
		return errors.Wrap(err, "failed to generate random seed in Round 1")
	}

	return stream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_Round1Response{
			Round1Response: &pb.Round1Response{
				Seed:      seed[:],
				Address:   msg.Address,
				SecretKey: msg.SecretKey,
			},
		},
	})
}

// func (h *SignHandler) handleRound3(stream pb.SignService_SignServer, alice *sign.Alice, msg *pb.Round2Response) error {
// 	fmt.Println("라운드3")
// 	round3Output, err := alice.Round3Sign(msg.Message, &sign.SignRound2Output{
// 		Seed: [32]byte(msg.Seed),
// 		DB:   msg.DB,
// 	})
// 	if err != nil {
// 		return errors.Wrap(err, "failed in Round3Sign")
// 	}
// 	return stream.Send(&pb.SignMessage{
// 		Msg: &pb.SignMessage_Round3Response{
// 			Round3Response: &pb.Round3Response{
// 				RSchnorrProof: round3Output.RSchnorrProof,
// 				RPrime:        round3Output.RPrime,
// 				EtaPhi:        round3Output.EtaPhi,
// 				EtaSig:        round3Output.EtaSig,
// 			},
// 		},
// 	})
// }
