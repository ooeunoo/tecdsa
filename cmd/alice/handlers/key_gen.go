package handlers

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"tecdsa/pkg/database/repository"
	deserializer "tecdsa/pkg/deserializers"
	"tecdsa/pkg/service"
	pb "tecdsa/proto/keygen"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/dkg"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

type keygenContext struct {
	network          int32
	alice            *dkg.Alice
	clientSecurityID uint32
	requestID        string
}

type KeygenHandler struct {
	curve          *curves.Curve
	repo           repository.ParitalSecretShareRepository
	networkService *service.NetworkService
}

func NewKeygenHandler(repo repository.ParitalSecretShareRepository, networkService *service.NetworkService) *KeygenHandler {
	return &KeygenHandler{
		curve:          curves.K256(),
		repo:           repo,
		networkService: networkService,
	}
}

func (h *KeygenHandler) HandleKeyGen(stream pb.KeygenService_KeyGenServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("no metadata received")
	}

	requestIDs := md.Get("request_id")
	if len(requestIDs) == 0 {
		return errors.New("request_id not found in metadata")
	}
	requestID := requestIDs[0]

	networkStr := md.Get("network")
	if len(networkStr) == 0 {
		return errors.New("network not found in metadata")
	}
	network, err := strconv.Atoi(networkStr[0])
	if err != nil {
		return errors.Wrap(err, "invalid network value in metadata")
	}

	clientSecurityIDStr := md.Get("client_security_id")
	if len(clientSecurityIDStr) == 0 {
		return errors.New("client_security_id not found in metadata")
	}
	clientSecurityID, err := strconv.ParseUint(clientSecurityIDStr[0], 10, 32)
	if err != nil {
		return errors.Wrap(err, "invalid client_security_id value in metadata")
	}

	ctx := &keygenContext{
		alice:            dkg.NewAlice(h.curve),
		requestID:        requestID,
		network:          int32(network),
		clientSecurityID: uint32(clientSecurityID),
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch msg := in.Msg.(type) {
		case *pb.KeygenMessage_KeyGenRound1To2Output:
			err = h.handleRound2(stream, ctx, msg.KeyGenRound1To2Output)
		case *pb.KeygenMessage_KeyGenRound3To4Output:
			err = h.handleRound4(stream, ctx, msg.KeyGenRound3To4Output)
		case *pb.KeygenMessage_KeyGenRound5To6Output:
			err = h.handleRound6(stream, ctx, msg.KeyGenRound5To6Output)
		case *pb.KeygenMessage_KeyGenRound7To8Output:
			err = h.handleRound8(stream, ctx, msg.KeyGenRound7To8Output)
		case *pb.KeygenMessage_KeyGenRound9To10Output:
			err = h.handleRound10(stream, ctx, msg.KeyGenRound9To10Output)
		default:
			err = fmt.Errorf("unexpected message type")
		}

		if err != nil {
			return err
		}
	}
}
func (h *KeygenHandler) handleRound2(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenRound1To2Output) error {
	fmt.Println("라운드2")

	round2Input, err := deserializer.DecodeDkgRound2Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 2")
	}

	round2Result, err := ctx.alice.Round2CommitToProof(round2Input)
	if err != nil {
		log.Printf("Error in Round2CommitToProof in Round 2")
		return err
	}

	roundPayload, err := deserializer.EncodeDkgRound2Output(round2Result)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 2")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound2To3Output{
			KeyGenRound2To3Output: &pb.KeyGenRound2To3Output{
				Payload: roundPayload,
			},
		},
	})
}

func (h *KeygenHandler) handleRound4(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenRound3To4Output) error {
	fmt.Println("라운드4")

	// msg
	payload := msg.Payload

	//
	round4Input, err := deserializer.DecodeDkgRound4Input(payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 4")
	}

	round4Result, err := ctx.alice.Round4VerifyAndReveal(round4Input)
	if err != nil {
		log.Printf("Error in Round4VerifyAndReveal in Round 4")
		return err
	}

	round4Payload, err := deserializer.EncodeDkgRound4Output(round4Result)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 4")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound4To5Output{
			KeyGenRound4To5Output: &pb.KeyGenRound4To5Output{
				Payload: round4Payload,
			},
		},
	})
}

func (h *KeygenHandler) handleRound6(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenRound5To6Output) error {
	fmt.Println("라운드6")

	// msg
	payload := msg.Payload

	//
	round6Input, err := deserializer.DecodeDkgRound6Input(payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 6")
	}

	round6Result, err := ctx.alice.Round6DkgRound2Ot(round6Input)
	if err != nil {
		return err
	}

	round6Payload, err := deserializer.EncodeDkgRound6Output(round6Result)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 6")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound6To7Output{
			KeyGenRound6To7Output: &pb.KeyGenRound6To7Output{
				Payload: round6Payload,
			},
		},
	})
}

func (h *KeygenHandler) handleRound8(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenRound7To8Output) error {
	fmt.Println("라운드8")

	// msg
	payload := msg.Payload

	//
	round8Input, err := deserializer.DecodeDkgRound8Input(payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 8")
	}

	round8Result, err := ctx.alice.Round8DkgRound4Ot(round8Input)
	if err != nil {
		return err
	}

	round8Payload, err := deserializer.EncodeDkgRound8Output(round8Result)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 8")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound8To9Output{
			KeyGenRound8To9Output: &pb.KeyGenRound8To9Output{
				Payload: round8Payload,
			},
		},
	})
}

func (h *KeygenHandler) handleRound10(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenRound9To10Output) error {
	fmt.Println("라운드10")

	round10Input, err := deserializer.DecodeDkgRound10Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 10")
	}

	roundErr := ctx.alice.Round10DkgRound6Ot(round10Input)
	if roundErr != nil {
		return roundErr
	}

	aliceOutput := ctx.alice.Output()
	networkObj, err := h.networkService.GetNetworkByID(ctx.network)
	if err != nil {
		return errors.Wrap(err, "failed to get network by ID")
	}

	address, err := h.networkService.DeriveAddress(aliceOutput.PublicKey, networkObj)
	if err != nil {
		return err
	}

	share, err := deserializer.EncodeAliceDkgOutput(aliceOutput)
	if err != nil {
		return errors.Wrap(err, "failed to encode alice output")
	}

	if err := h.repo.Create(address, share, uint(ctx.clientSecurityID)); err != nil {
		return errors.Wrap(err, "failed to store secret alice share")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound10To11Output{
			KeyGenRound10To11Output: &pb.KeyGenRound10To11Output{},
		},
	})
}
