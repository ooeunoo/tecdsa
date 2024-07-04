package handlers

import (
	"context"
	"log"
	"strconv"

	"tecdsa/internal/encoding"
	pb "tecdsa/pkg/api/grpc/dkg"

	"google.golang.org/grpc"
)

type DKGHandler struct {
	pb.UnimplementedDKGServiceServer
}

func (h *DKGHandler) ProcessDKG(ctx context.Context, req *pb.DKGRequest) (*pb.DKGResponse, error) {
	// 랜덤 시드 생성
	seed, _ := strconv.Atoi(req.RandomNumber)
	encodedSeed := encoding.Encode(seed)

	// Alice에게 gRPC 요청 보내기
	conn, err := grpc.Dial("alice:50052", grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to Alice: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewDKGServiceClient(conn)
	resp, err := client.ProcessDKG(ctx, &pb.DKGRequest{RandomNumber: encodedSeed})
	if err != nil {
		log.Printf("Error calling Alice's ProcessDKG: %v", err)
		return nil, err
	}

	// Alice로부터 받은 결과 처리
	decodedResult := encoding.Decode(resp.Result)
	log.Printf("Received from Alice: %d", decodedResult)

	return &pb.DKGResponse{Result: encoding.Encode(decodedResult)}, nil
}
