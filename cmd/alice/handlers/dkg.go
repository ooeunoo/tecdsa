package server

// import (
// 	"io"

// 	"your_project/internal/dkls/dkg"
// 	pb "your_project/pkg/api/grpc/dkg"

// 	"github.com/coinbase/kryptology/pkg/core/curves"
// )

// type Server struct {
// 	pb.UnimplementedDkgServiceServer
// 	alice *dkg.Alice
// }

// func NewServer() *Server {
// 	curve := curves.K256()
// 	return &Server{
// 		alice: dkg.NewAlice(curve),
// 	}
// }

// func (s *Server) KeyGen(stream pb.DkgService_KeyGenServer) error {
// 	for {
// 		req, err := stream.Recv()
// 		if err == io.EOF {
// 			return nil
// 		}
// 		if err != nil {
// 			return err
// 		}

// 		switch msg := req.Msg.(type) {
// 		case *pb.DkgMessage_Round1Request:
// 			// Alice doesn't generate a seed in Round 1, so we return an empty response
// 			if err := stream.Send(&pb.DkgMessage{Msg: &pb.DkgMessage_Round1Response{Round1Response: &pb.Round1Response{}}}); err != nil {
// 				return err
// 			}
// 		case *pb.DkgMessage_Round2Request:
// 			round2Output, err := s.alice.Round2CommitToProof([32]byte(msg.Round2Request.BobSeed))
// 			if err != nil {
// 				return err
// 			}
// 			if err := stream.Send(&pb.DkgMessage{Msg: &pb.DkgMessage_Round2Response{Round2Response: &pb.Round2Response{AliceSeed: round2Output.Seed[:], Commitment: round2Output.Commitment}}}); err != nil {
// 				return err
// 			}
// 			// Handle other rounds similarly
// 		}
// 	}
// }
