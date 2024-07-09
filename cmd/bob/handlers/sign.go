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
}

func NewSignHandler(repo repository.SecretRepository) *SignHandler {
	return &SignHandler{
		curve: curves.K256(),
		hash:  sha3.NewLegacyKeccak256(),
		repo:  repo,
	}
}

func (h *SignHandler) HandleSign(stream pb.SignService_SignServer) error {

	var bob *sign.Bob // Bob 인스턴스를 저장할 변수

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch msg := in.Msg.(type) {
		case *pb.SignMessage_Round1Response:
			err = h.handleRound2(stream, msg.Round1Response)
		case *pb.SignMessage_Round3Response:
			err = h.handleRound4(stream, bob, msg.Round3Response)
		default:
			err = fmt.Errorf("unexpected message type")
		}

		if err != nil {
			return err
		}
	}
}

func (h *SignHandler) handleRound2(stream pb.SignService_SignServer, msg *pb.Round1Response) error {
	fmt.Println("라운드2")

	// GetSecretShare 함수를 사용하여 AliceOutput 가져오기
	bobOutputInterface, err := h.repo.GetSecretShare(msg.Address, msg.SecretKey)
	if err != nil {
		return errors.Wrap(err, "failed to get secret share")
	}

	bobOutput, ok := bobOutputInterface.(*dkg.BobOutput)
	if !ok {
		return errors.New("retrieved secret share is not an BobOutput")
	}

	// bob 인스턴스 생성
	bob := sign.NewBob(h.curve, h.hash, bobOutput)

	fmt.Println("라운드2")
	round2Output, err := bob.Round2Initialize([32]byte(msg.Seed))
	if err != nil {
		return errors.Wrap(err, "failed in Round2Initialize")
	}

	kosRound1Outputs := make([]*pb.KosRound1Output, len(round2Output.KosRound1Outputs))
	for i, output := range round2Output.KosRound1Outputs {
		u := make([][]byte, len(output.U))
		for j, row := range output.U {
			u[j] = row[:]
		}
		kosRound1Outputs[i] = &pb.KosRound1Output{
			U:      u,
			WPrime: output.WPrime[:],
			VPrime: output.VPrime[:],
		}
	}

	return stream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_Round2Response{
			Round2Response: &pb.Round2Response{
				Output: &pb.SignRound2Output{
					KosRound1Outputs: kosRound1Outputs,
					Db:               round2Output.DB.ToAffineCompressed(),
					Seed:             round2Output.Seed[:],
				},
			},
		},
	})
}

// func (h *SignHandler) handleRound2(stream pb.SignService_SignServer, bob *sign.Bob, msg *pb.Round1Response) error {
// 	fmt.Println("라운드1")

// 	// GetSecretShare 함수를 사용하여 AliceOutput 가져오기
// 	aliceOutputInterface, err := h.repo.GetSecretShare(msg.Address, msg.SecretKey)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get secret share")
// 	}

// 	aliceOutput, ok := aliceOutputInterface.(*dkg.BobOutput)
// 	if !ok {
// 		return errors.New("retrieved secret share is not an AliceOutput")
// 	}

// 	// Alice 인스턴스 생성
// 	bob := sign.NewBob(h.curve, h.hash, aliceOutput)

// 	fmt.Println("라운드2")
// 	round2Output, err := bob.Round2Initialize([32]byte(msg.Seed), msg.Message)
// 	if err != nil {
// 		return errors.Wrap(err, "failed in Round2Initialize")
// 	}
// 	return stream.Send(&pb.SignMessage{
// 		Msg: &pb.SignMessage_Round2Response{
// 			Round2Response: &pb.Round2Response{
// 				Seed: round2Output.Seed[:],
// 				DB:   round2Output.DB,
// 			},
// 		},
// 	})
// }

func (h *SignHandler) handleRound4(stream pb.SignService_SignServer, bob *sign.Bob, msg *pb.Round3Response) error {
	fmt.Println("라운드4")
	err := bob.Round4Final(msg.Message, &sign.SignRound3Output{
		RSchnorrProof: msg.RSchnorrProof,
		RPrime:        msg.RPrime,
		EtaPhi:        msg.EtaPhi,
		EtaSig:        msg.EtaSig,
	})
	if err != nil {
		return errors.Wrap(err, "failed in Round4Final")
	}
	return stream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_FinalResponse{
			FinalResponse: &pb.FinalResponse{
				Signature: bob.Signature,
			},
		},
	})
}
