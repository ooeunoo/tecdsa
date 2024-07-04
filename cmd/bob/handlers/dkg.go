package handlers

import (
	"context"

	encoding "tecdsa/internal/encoding/dkg"

	"github.com/coinbase/kryptology/pkg/core/curves"
	v1 "github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1"

	dkg "github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/dkg"

	pb "tecdsa/pkg/api/grpc/dkg"
)

type DkgHandler struct {
	pb.UnimplementedDkgServiceServer
}

func NewDkgHandler() *DkgHandler {
	return &DkgHandler{}
}

func (h *DkgHandler) GenerateDkg(ctx context.Context, req *pb.DkgRequest) (*pb.DkgResponse, error) {
	curve := curves.K256()

	bob := dkg.NewBob(curve)

	// Implement DKG rounds for Bob
	seed, err := bob.Round1GenerateRandomSeed()
	if err != nil {
		return nil, err
	}

	// Encode seed to send to Alice
	encodedSeed, err := v1.EncodingDkgRound(seed)
	if err != nil {
		return nil, err
	}

	// Send encodedSeed to Alice and receive encodedRound2Input
	// This step would involve actual network communication in a real implementation

	// Decode Alice's response
	round2Input, err := encoding.DecodeDkgRound2Input(encodedRound2Input)
	if err != nil {
		return nil, err
	}

	proof, err := bob.Round3SchnorrProve(round2Input)
	if err != nil {
		return nil, err
	}

	// Continue with remaining rounds...

	// After completing DKG rounds:
	output := bob.Output()

	// Encode the final output
	encodedOutput, err := encoding.EncodeBobDkgOutput(output)
	if err != nil {
		return nil, err
	}

	return &pb.DkgResponse{
		SessionId:     req.SessionId,
		EncodedOutput: encodedOutput,
	}, nil
}
