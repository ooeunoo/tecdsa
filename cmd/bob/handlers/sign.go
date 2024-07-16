package handlers

import (
	"encoding/base64"
	"encoding/hex"
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
	bob       *sign.Bob
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
	fmt.Println("addresses:", addresses)
	if len(addresses) == 0 {
		return errors.New("address not found in metadata")
	}
	address := addresses[0]

	txOrigins := md.Get("tx_origin")
	if len(txOrigins) == 0 {
		return errors.New("tx_origin not found in metadata")
	}
	txOrigin, err := base64.StdEncoding.DecodeString(txOrigins[0])
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
		case *pb.SignMessage_SignRound1To2Output:
			err = h.handleRound2(stream, ctx, msg.SignRound1To2Output)
		case *pb.SignMessage_SignRound3To4Output:
			err = h.handleRound4(stream, ctx, msg.SignRound3To4Output)
		default:
			err = errors.New("unexpected message type")
		}

		if err != nil {
			log.Printf("Error in signing process: %v", err)
			return err
		}
	}
}

func (h *SignHandler) handleRound2(stream pb.SignService_SignServer, ctx *signContext, msg *pb.SignRound1To2Output) error {
	log.Printf("라운드2")

	output, err := h.repo.FindByAddress(ctx.address)
	if err != nil {
		return errors.Wrap(err, "failed to get secret share")
	}

	bobOutput, err := deserializer.DecodeBobDkgResult(output.Share)
	if err != nil {
		return errors.New("retrieved secret share is not a BobOutput")
	}

	publicKeyBytes := bobOutput.PublicKey.ToAffineCompressed()
	publicKeyHex := hex.EncodeToString(publicKeyBytes)
	fmt.Println("Public Key (hex):", publicKeyHex)
	ctx.bob = sign.NewBob(h.curve, h.hash, bobOutput)

	round1Payload, err := deserializer.DecodeSignRound1Payload(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode round 1 payload")
	}

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
	log.Printf("라운드4")

	round3Payload, err := deserializer.DecodeSignRound3Payload(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 4")
	}

	fmt.Println("bob txOrigin:", ctx.txOrigin)
	if err = ctx.bob.Round4Final(ctx.txOrigin, round3Payload); err != nil {
		return errors.Wrap(err, "failed in Round4Final")
	}

	signature := ctx.bob.Signature
	return stream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_SignRound4ToGatewayOutput{
			SignRound4ToGatewayOutput: &pb.SignRound4ToGatewayOutput{
				V:         uint64(signature.V),
				R:         signature.R.Bytes(),
				S:         signature.S.Bytes(),
				RequestId: ctx.requestID,
			},
		},
	})
}
