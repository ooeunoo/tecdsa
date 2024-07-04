package handlers

import (
	"context"

	"tecdsa/internal/encoding"
	pb "tecdsa/pkg/api/grpc/dkg"
)

type DKGHandler struct {
	pb.UnimplementedDKGServiceServer
}

func (h *DKGHandler) ProcessDKG(ctx context.Context, req *pb.DKGRequest) (*pb.DKGResponse, error) {
	decodedNumber := encoding.Decode(req.RandomNumber)
	result := decodedNumber * 3
	encodedResult := encoding.Encode(result)

	return &pb.DKGResponse{Result: encodedResult}, nil
}
