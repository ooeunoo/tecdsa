package server

import (
	"fmt"
	"io"
	"tecdsa/internal/dkls/dkg"
	pb "tecdsa/pkg/api/grpc/dkg"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/ot/base/simplest"
	"github.com/coinbase/kryptology/pkg/zkp/schnorr"
)

type Server struct {
	pb.UnimplementedDkgServiceServer
	alice *dkg.Alice
}

func NewServer() *Server {
	curve := curves.K256()
	return &Server{
		alice: dkg.NewAlice(curve),
	}
}

func (s *Server) KeyGen(stream pb.DkgService_KeyGenServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch msg := in.Msg.(type) {
		// Round 2
		case *pb.DkgMessage_Round1Response:
			fmt.Println("라운드2")

			round2Output, err := s.alice.Round2CommitToProof([32]byte(msg.Round1Response.Seed))
			if err != nil {
				return err
			}
			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_Round2Response{
					Round2Response: &pb.Round2Response{
						Seed:       round2Output.Seed[:],
						Commitment: round2Output.Commitment,
					},
				},
			}); err != nil {
				return err
			}
			// Round 4
		case *pb.DkgMessage_Round3Response:
			fmt.Println("라운드4")

			k256 := curves.K256()
			c, err := k256.Scalar.SetBytes(msg.Round3Response.C)
			if err != nil {
				return err
			}
			S, err := k256.Scalar.SetBytes(msg.Round3Response.S)
			if err != nil {
				return err
			}
			statement, err := k256.Point.FromAffineCompressed(msg.Round3Response.Statement)
			if err != nil {
				return err
			}
			schnorrProof := &schnorr.Proof{
				C:         c,
				S:         S,
				Statement: statement,
			}
			proof, err := s.alice.Round4VerifyAndReveal(schnorrProof)
			if err != nil {
				return err
			}
			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_Round4Response{
					Round4Response: &pb.Round4Response{
						C:         proof.C.Bytes(),
						S:         proof.S.Bytes(),
						Statement: proof.Statement.ToAffineCompressed(),
					},
				},
			}); err != nil {
				return err
			}
			// Round 6
		case *pb.DkgMessage_Round5Response:
			fmt.Println("라운드6")

			k256 := curves.K256()
			c, err := k256.Scalar.SetBytes(msg.Round5Response.C)
			if err != nil {
				return err
			}
			S, err := k256.Scalar.SetBytes(msg.Round5Response.S)
			if err != nil {
				return err
			}
			statement, err := k256.Point.FromAffineCompressed(msg.Round5Response.Statement)
			if err != nil {
				return err
			}
			schnorrProof := &schnorr.Proof{
				C:         c,
				S:         S,
				Statement: statement,
			}
			compressedReceiversMaskedChoice, err := s.alice.Round6DkgRound2Ot(schnorrProof)
			if err != nil {
				return err
			}
			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_Round6Response{
					Round6Response: &pb.Round6Response{
						ReceiversMaskedChoices: compressedReceiversMaskedChoice,
					},
				},
			}); err != nil {
				return err
			}

			// Round8
		case *pb.DkgMessage_Round7Response:
			fmt.Println("라운드8")

			challenges := make([]simplest.OtChallenge, len(msg.Round7Response.OtChallenges))
			for i, c := range msg.Round7Response.OtChallenges {
				copy(challenges[i][:], c)
			}
			challengeResponse, err := s.alice.Round8DkgRound4Ot(challenges)
			if err != nil {
				return err
			}
			challengeResponseBytes := make([][]byte, len(challengeResponse))
			for i, cr := range challengeResponse {
				challengeResponseBytes[i] = cr[:]
			}
			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_Round8Response{
					Round8Response: &pb.Round8Response{
						OtChallengeResponses: challengeResponseBytes,
					},
				},
			}); err != nil {
				return err
			}

			// Round10
		case *pb.DkgMessage_Round9Response:
			fmt.Println("라운드10")

			challengeOpenings := make([]simplest.ChallengeOpening, len(msg.Round9Response.ChallengeOpenings))
			for i, co := range msg.Round9Response.ChallengeOpenings {
				// co는 []byte 타입이므로, 이를 [2][32]byte 타입으로 변환
				if len(co) != 2*32 {
					return fmt.Errorf("invalid challenge opening length")
				}
				copy(challengeOpenings[i][0][:], co[:32])
				copy(challengeOpenings[i][1][:], co[32:])
			}
			err := s.alice.Round10DkgRound6Ot(challengeOpenings)
			if err != nil {
				return err
			}

			secretKeyShare := s.alice.Output().SecretKeyShare.Bytes()
			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_Round10Response{
					Round10Response: &pb.Round10Response{
						Success:             true,
						AliceSecretKeyShare: secretKeyShare,
					},
				},
			}); err != nil {
				return err
			}
		}
	}
}
