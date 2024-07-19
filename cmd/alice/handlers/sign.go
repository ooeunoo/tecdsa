package handlers

import (
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"log"
	"tecdsa/pkg/database/repository"
	deserializer "tecdsa/pkg/deserializers"
	pb "tecdsa/proto/sign"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/sign"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc/metadata"
)

type signContext struct {
	alice     *sign.Alice
	txOrigin  []byte
	requestID string
	address   string
}

type SignHandler struct {
	curve *curves.Curve
	hash  hash.Hash
	repo  repository.ParitalSecretShareRepository
}

func NewSignHandler(repo repository.ParitalSecretShareRepository) *SignHandler {
	return &SignHandler{
		curve: curves.K256(),
		hash:  sha3.NewLegacyKeccak256(),
		repo:  repo,
	}
}

func (h *SignHandler) HandleSign(stream pb.SignService_SignServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("no metadata received")
	}

	requestIDs := md.Get("request_id")
	if len(requestIDs) == 0 {
		return errors.New("request_id not found in metadata")
	}
	requestID := requestIDs[0]

	addresses := md.Get("address")
	if len(addresses) == 0 {
		return errors.New("address not found in metadata")
	}
	address := addresses[0]

	txOrigins := md.Get("tx_origin")
	if len(txOrigins) == 0 {
		return errors.New("tx_origin not found in metadata")
	}
	txOrigin, err := base64.StdEncoding.DecodeString(txOrigins[0])
	fmt.Println("txOrigin:", txOrigin)
	if err != nil {
		return errors.Wrap(err, "failed to decode tx_origin")
	}
	ctx := &signContext{
		requestID: requestID,
		address:   address,
		txOrigin:  txOrigin,
	}

	log.Printf("Starting signing process for request ID: %s, address: %s", requestID, address)

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "error receiving message")
		}

		switch msg := in.Msg.(type) {
		case *pb.SignMessage_SignGatewayTo1Output:
			err = h.handleRound1(stream, ctx, msg.SignGatewayTo1Output)
		case *pb.SignMessage_SignRound2To3Output:
			err = h.handleRound3(stream, ctx, msg.SignRound2To3Output)
		default:
			err = errors.New("unexpected message type")
		}

		if err != nil {
			log.Printf("Error in signing process: %v", err)
			return err
		}
	}
}

func (h *SignHandler) handleRound1(stream pb.SignService_SignServer, ctx *signContext, msg *pb.SignGatewayTo1Output) error {
	log.Printf("라운드1")

	output, err := h.repo.FindByAddress(ctx.address)
	if err != nil {
		return errors.Wrap(err, "failed to get secret share")
	}

	aliceOutput, err := deserializer.DecodeAliceDkgResult(output.Share)
	if err != nil {
		return errors.New("retrieved secret share is not an AliceOutput")
	}

	ctx.alice = sign.NewAlice(h.curve, h.hash, aliceOutput)

	round1Result, err := ctx.alice.Round1GenerateRandomSeed()
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
				Payload: round1Payload,
			},
		},
	})
}

func (h *SignHandler) handleRound3(stream pb.SignService_SignServer, ctx *signContext, msg *pb.SignRound2To3Output) error {
	log.Printf("라운드3")

	round2Payload, err := deserializer.DecodeSignRound2Payload(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 3 input")
	}

	round3Result, err := ctx.alice.Round3Sign(ctx.txOrigin, round2Payload)
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
				Payload: round3Payload,
			},
		},
	})
}
