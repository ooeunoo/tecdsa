package handlers

import (
	"context"

	"tecdsa/internal/dkg"
	encoding "tecdsa/internal/encoding/dkg"

	pb "tecdsa/pkg/api/grpc/dkg"
)

type DkgHandler struct {
	pb.UnimplementedDkgServiceServer
}

func NewDkgHandler() *DkgHandler {
	return &DkgHandler{}
}

func (h *DkgHandler) GenerateDkg(ctx context.Context, req *pb.DkgRequest) (*pb.DkgResponse, error) {
	alice := dkg.NewAlice(dkg.DefaultCurve())

	// Implement DKG rounds for Alice
	seed, err := alice.Round1GenerateRandomSeed()
	if err != nil {
		return nil, err
	}

	round2Output, err := alice.Round2CommitToProof(seed)
	if err != nil {
		return nil, err
	}

	// Encode round2Output to send to Bob
	encodedRound2Output, err := encoding.EncodeDkgRound2Output(round2Output)
	if err != nil {
		return nil, err
	}

	// Send encodedRound2Output to Bob and receive encodedRound3Input
	// This step would involve actual network communication in a real implementation

	// Decode Bob's response
	round3Input, err := encoding.DecodeDkgRound3Input(encodedRound3Input)
	if err != nil {
		return nil, err
	}

	proof, err := alice.Round4VerifyAndReveal(round3Input)
	if err != nil {
		return nil, err
	}

	// Continue with remaining rounds...

	// After completing DKG rounds:
	output := alice.Output()

	// Encode the final output
	encodedOutput, err := encoding.EncodeAliceDkgOutput(output)
	if err != nil {
		return nil, err
	}

	return &pb.DkgResponse{
		SessionId:     req.SessionId,
		EncodedOutput: encodedOutput,
	}, nil
}
