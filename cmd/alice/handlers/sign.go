// alice handler
package handlers

import (
	"fmt"
	"hash"
	"io"
	"tecdsa/pkg/database/repository"
	deserializer "tecdsa/pkg/deserializers"
	pb "tecdsa/proto/sign"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/sign"
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
		case *pb.SignMessage_SignRequestTo1Output:
			err = h.handleRound1(stream, msg.SignRequestTo1Output)
		case *pb.SignMessage_SignRound2To3Output:
			err = h.handleRound3(stream, msg.SignRound2To3Output)
		default:
			err = fmt.Errorf("unexpected message type")
		}

		if err != nil {
			return err
		}
	}
}

func (h *SignHandler) handleRound1(stream pb.SignService_SignServer, msg *pb.SignRequestTo1Output) error {
	fmt.Println("라운드1")

	// msg
	payload := msg.Payload

	params, err := deserializer.DecodeSignRequestToRound1(payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 1")
	}

	// destructors
	address := params.Address
	secretKey := params.SecretKey

	//
	output, err := h.repo.GetSecretShare(address, secretKey)
	if err != nil {
		return errors.Wrap(err, "failed to get secret share")
	}

	aliceOutput, err := deserializer.DecodeAliceDkgResult(output)
	if err != nil {
		return errors.New("retrieved secret share is not an AliceOutput")
	}

	h.alice = sign.NewAlice(h.curve, h.hash, aliceOutput)

	//
	round1Result, err := h.alice.Round1GenerateRandomSeed()
	if err != nil {
		return errors.Wrap(err, "failed to generate random seed in Round 1")
	}

	round1Payload, err := deserializer.EncodeSignRound1Payload(round1Result)
	if err != nil {
		return errors.Wrap(err, "failed to encode result in Round 1")
	}

	return stream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_SignRound1To2Output{
			SignRound1To2Output: &pb.SignRound1To2Output{
				Address:   params.Address,
				SecretKey: params.SecretKey,
				TxOrigin:  params.TxOrigin,
				Payload:   round1Payload,
			},
		},
	})
}

func (h *SignHandler) handleRound3(stream pb.SignService_SignServer, msg *pb.SignRound2To3Output) error {
	fmt.Println("라운드3")

	// msg
	payload := msg.Payload
	txOrigin := msg.TxOrigin

	round2Payload, err := deserializer.DecodeSignRound2Payload(payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 3 input")
	}

	round3Result, err := h.alice.Round3Sign(txOrigin, round2Payload)
	if err != nil {
		return errors.Wrap(err, "failed to sign in Round 3")
	}

	round3Payload, err := deserializer.EncodeSignRound3Payload(round3Result)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 3 payload")
	}

	return stream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_SignRound3To4Output{
			SignRound3To4Output: &pb.SignRound3To4Output{
				TxOrigin: txOrigin,
				Payload:  round3Payload,
			},
		},
	})

}
