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

type signContext struct {
	bob      *sign.Bob
	txOrigin []byte
}

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
	ctx := &signContext{}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch msg := in.Msg.(type) {
		case *pb.SignMessage_SignRound1To2Output:
			err = h.handleRound2(stream, ctx, msg.SignRound1To2Output)
		case *pb.SignMessage_SignRound3To4Output:
			err = h.handleRound4(stream, ctx, msg.SignRound3To4Output)
		default:
			err = fmt.Errorf("unexpected message type")
		}

		if err != nil {
			return err
		}
	}
}

func (h *SignHandler) handleRound2(stream pb.SignService_SignServer, ctx *signContext, msg *pb.SignRound1To2Output) error {
	fmt.Println("라운드2")

	// param
	payload := msg.Payload
	address := msg.Address
	secretKey := msg.SecretKey
	ctx.txOrigin = msg.TxOrigin

	round1Payload, err := deserializer.DecodeSignRound1Payload(payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode round 1 payload")
	}
	fmt.Println("address:", address)
	fmt.Println("secretKey:", secretKey)

	output, err := h.repo.GetSecretShare(address, secretKey)
	if err != nil {
		return errors.Wrap(err, "failed to get secret share")
	}

	bobOutput, err := deserializer.DecodeBobDkgResult(output)
	if err != nil {
		return errors.New("retrieved secret share is not a BobOutput")
	}

	ctx.bob = sign.NewBob(h.curve, h.hash, bobOutput)

	round2Result, err := ctx.bob.Round2Initialize(round1Payload)
	if err != nil {
		return errors.Wrap(err, "failed in Round2Initialize")
	}

	round2Payload, err := deserializer.EncodeSignRound2Payload(round2Result)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 2")
	}

	return stream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_SignRound2To3Output{
			SignRound2To3Output: &pb.SignRound2To3Output{
				Payload: round2Payload,
			},
		},
	})
}

func (h *SignHandler) handleRound4(stream pb.SignService_SignServer, ctx *signContext, msg *pb.SignRound3To4Output) error {
	fmt.Println("라운드4")

	// msg
	payload := msg.Payload

	//
	round3Payload, err := deserializer.DecodeSignRound3Payload(payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 4")
	}

	if err = ctx.bob.Round4Final(ctx.txOrigin, round3Payload); err != nil {
		return errors.Wrap(err, "failed in Round4Final")
	}

	signature := ctx.bob.Signature
	return stream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_SignRound4ToResponseOutput{
			SignRound4ToResponseOutput: &pb.SignRound4ToResponseOutput{
				V: uint64(signature.V),
				R: signature.R.Bytes(),
				S: signature.S.Bytes(),
			},
		},
	})
}
