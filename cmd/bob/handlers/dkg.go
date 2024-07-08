package handlers

// import (
// 	"context"

// 	"tecdsa/internal/dkls/dkg"
// 	pb "tecdsa/pkg/api/grpc/dkg"

// 	"google.golang.org/grpc"
// )

// func ProcessKeyGen(ctx context.Context, bob *dkg.Bob) (*pb.KeyGenResponse, error) {
// 	conn, err := grpc.Dial("alice:50052", grpc.WithInsecure())
// 	if err != nil {
// 		return &pb.KeyGenResponse{Success: false}, err
// 	}
// 	defer conn.Close()

// 	client := pb.NewDkgServiceClient(conn)

// 	// Round 1
// 	bobSeed, err := bob.Round1GenerateRandomSeed()
// 	if err != nil {
// 		return &pb.KeyGenResponse{Success: false}, err
// 	}
// 	round1Resp, err := client.Round1GenerateRandomSeed(ctx, &pb.Empty{})
// 	if err != nil {
// 		return &pb.KeyGenResponse{Success: false}, err
// 	}

// 	// Round 2
// 	round2Resp, err := client.Round2CommitToProof(ctx, &pb.Round1Response{BobSeed: bobSeed[:]})
// 	if err != nil {
// 		return &pb.KeyGenResponse{Success: false}, err
// 	}

// 	// Round 3
// 	round2Output := &dkg.Round2Output{
// 		Seed:       round2Resp.Seed,
// 		Commitment: round2Resp.Commitment,
// 	}
// 	proof, err := bob.Round3SchnorrProve(round2Output)
// 	if err != nil {
// 		return &pb.KeyGenResponse{Success: false}, err
// 	}
// 	round3Resp, err := client.Round3SchnorrProve(ctx, &pb.Round3Request{
// 		C:         proof.C.Bytes(),
// 		S:         proof.S.Bytes(),
// 		Statement: proof.Statement.ToAffineCompressed(),
// 	})
// 	if err != nil {
// 		return &pb.KeyGenResponse{Success: false}, err
// 	}

// 	// Round 4
// 	aliceProof := &dkg.Proof{
// 		C:         bob.curve.Scalar.FromBytes(round3Resp.C),
// 		S:         bob.curve.Scalar.FromBytes(round3Resp.S),
// 		Statement: bob.curve.Point.FromAffineCompressed(round3Resp.Statement),
// 	}
// 	bobProof, err := bob.Round5DecommitmentAndStartOt(aliceProof)
// 	if err != nil {
// 		return &pb.KeyGenResponse{Success: false}, err
// 	}
// 	round5Resp, err := client.Round5DecommitmentAndStartOt(ctx, &pb.Round5Request{
// 		C:         bobProof.C.Bytes(),
// 		S:         bobProof.S.Bytes(),
// 		Statement: bobProof.Statement.ToAffineCompressed(),
// 	})
// 	if err != nil {
// 		return &pb.KeyGenResponse{Success: false}, err
// 	}

// 	// Rounds 6-10 (OT rounds)
// 	// Implement the OT rounds here, similar to the above rounds

// 	// Final output
// 	output := bob.Output()
// 	return &pb.KeyGenResponse{Success: true}, nil
// }
