package handlers

import (
	"fmt"
	"io"
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
	bob              *dkg.Bob
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
		bob:              dkg.NewBob(h.curve),
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
		case *pb.KeygenMessage_KeyGenGatewayTo1Output:
			err = h.handleRound1(stream, ctx, msg.KeyGenGatewayTo1Output)
		case *pb.KeygenMessage_KeyGenRound2To3Output:
			err = h.handleRound3(stream, ctx, msg.KeyGenRound2To3Output)
		case *pb.KeygenMessage_KeyGenRound4To5Output:
			err = h.handleRound5(stream, ctx, msg.KeyGenRound4To5Output)
		case *pb.KeygenMessage_KeyGenRound6To7Output:
			err = h.handleRound7(stream, ctx, msg.KeyGenRound6To7Output)
		case *pb.KeygenMessage_KeyGenRound8To9Output:
			err = h.handleRound9(stream, ctx, msg.KeyGenRound8To9Output)
		case *pb.KeygenMessage_KeyGenRound10To11Output:
			err = h.handleRound11(stream, ctx, msg.KeyGenRound10To11Output)
		default:
			err = fmt.Errorf("unexpected message type")
		}

		if err != nil {
			return err
		}
	}
}

func (h *KeygenHandler) handleRound1(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenGatewayTo1Output) error {
	fmt.Println("라운드1")

	seed, err := ctx.bob.Round1GenerateRandomSeed()
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

func (h *KeygenHandler) handleRound3(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenRound2To3Output) error {
	fmt.Println("라운드3")

	round3Input, err := deserializer.DecodeDkgRound3Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 3")
	}

	proof, err := ctx.bob.Round3SchnorrProve(round3Input)
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

func (h *KeygenHandler) handleRound5(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenRound4To5Output) error {
	fmt.Println("라운드5")

	round5Input, err := deserializer.DecodeDkgRound5Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 5")
	}

	proof, err := ctx.bob.Round5DecommitmentAndStartOt(round5Input)
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

func (h *KeygenHandler) handleRound7(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenRound6To7Output) error {
	fmt.Println("라운드7")

	round7Input, err := deserializer.DecodeDkgRound7Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 7")
	}

	challenges, err := ctx.bob.Round7DkgRound3Ot(round7Input)
	if err != nil {
		return errors.Wrap(err, "failed in Round7DkgRound3Ot in Round 7")
	}

	round7Output, err := deserializer.EncodeDkgRound7Output(challenges)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 7")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound7To8Output{
			KeyGenRound7To8Output: &pb.KeyGenRound7To8Output{
				Payload: round7Output,
			},
		},
	})
}

func (h *KeygenHandler) handleRound9(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenRound8To9Output) error {
	fmt.Println("라운드9")
	round9Input, err := deserializer.DecodeDkgRound9Input(msg.Payload)
	if err != nil {
		return errors.Wrap(err, "failed to decode in Round 9")
	}

	challengeOpenings, err := ctx.bob.Round9DkgRound5Ot(round9Input)
	if err != nil {
		return errors.Wrap(err, "failed in Round9DkgRound5Ot in Round 9")
	}

	round9Output, err := deserializer.EncodeDkgRound9Output(challengeOpenings)
	if err != nil {
		return errors.Wrap(err, "failed to encode in Round 9")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound9To10Output{
			KeyGenRound9To10Output: &pb.KeyGenRound9To10Output{
				Payload: round9Output,
			}},
	})
}

func (h *KeygenHandler) handleRound11(stream pb.KeygenService_KeyGenServer, ctx *keygenContext, msg *pb.KeyGenRound10To11Output) error {
	fmt.Println("라운드끝")

	bobOutput := ctx.bob.Output()
	networkObj, err := h.networkService.GetNetworkByID(ctx.network)
	if err != nil {
		return errors.Wrap(err, "failed to get network by ID")
	}

	address, err := h.networkService.DeriveAddress(bobOutput.PublicKey, networkObj)
	if err != nil {
		return err
	}

	share, err := deserializer.EncodeBobDkgOutput(bobOutput)
	if err != nil {
		return errors.Wrap(err, "failed to encode bob output")
	}

	if err := h.repo.Create(address, share, uint(ctx.clientSecurityID)); err != nil {
		return errors.Wrap(err, "failed to store secret bob share")
	}

	return stream.Send(&pb.KeygenMessage{
		Msg: &pb.KeygenMessage_KeyGenRound11ToGatewayOutput{
			KeyGenRound11ToGatewayOutput: &pb.KeyGenRound11ToGatewayOutput{
				Address:   address,
				RequestId: ctx.requestID,
			},
		},
	})
}
