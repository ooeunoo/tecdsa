package server

import (
	"fmt"
	"io"
	"log"
	"tecdsa/internal/dkls/dkg"
	pb "tecdsa/pkg/api/grpc/dkg"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/ot/base/simplest"
	"github.com/coinbase/kryptology/pkg/zkp/schnorr"
	"github.com/ethereum/go-ethereum/crypto"
)

type Server struct {
	pb.UnimplementedDkgServiceServer
	bob *dkg.Bob
}

func NewServer() *Server {
	curve := curves.K256()
	return &Server{
		bob: dkg.NewBob(curve),
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
		// Round1
		case *pb.DkgMessage_Round1Request:
			fmt.Println("라운드1")
			seed, err := s.bob.Round1GenerateRandomSeed()
			if err != nil {
				return err
			}
			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_Round1Response{
					Round1Response: &pb.Round1Response{Seed: seed[:]},
				},
			}); err != nil {
				return err
			}
		// Round3
		case *pb.DkgMessage_Round2Response:
			fmt.Println("라운드3")

			proof, err := s.bob.Round3SchnorrProve(&dkg.Round2Output{
				Seed:       [32]byte(msg.Round2Response.Seed),
				Commitment: msg.Round2Response.Commitment,
			})
			if err != nil {
				return err
			}
			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_Round3Response{
					Round3Response: &pb.Round3Response{
						C:         proof.C.Bytes(),
						S:         proof.S.Bytes(),
						Statement: proof.Statement.ToAffineCompressed(),
					},
				},
			}); err != nil {
				return err
			}
			// Round 5
		case *pb.DkgMessage_Round4Response:
			fmt.Println("라운드5")

			k256 := curves.K256()
			c, err := k256.Scalar.SetBytes(msg.Round4Response.C)
			if err != nil {
				return err
			}
			S, err := k256.Scalar.SetBytes(msg.Round4Response.S)
			if err != nil {
				return err
			}
			statement, err := k256.Point.FromAffineCompressed(msg.Round4Response.Statement)
			if err != nil {
				return err
			}

			schnorrProof := &schnorr.Proof{
				C:         c,
				S:         S,
				Statement: statement,
			}

			proof, err := s.bob.Round5DecommitmentAndStartOt(schnorrProof)
			if err != nil {
				return err
			}
			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_Round5Response{
					Round5Response: &pb.Round5Response{
						C:         proof.C.Bytes(),
						S:         proof.S.Bytes(),
						Statement: proof.Statement.ToAffineCompressed(),
					}},
			}); err != nil {
				return err
			}
			// Round 7

		case *pb.DkgMessage_Round6Response:
			fmt.Println("라운드7")

			compressedReceiversMaskedChoice := make([]simplest.ReceiversMaskedChoices, len(msg.Round6Response.ReceiversMaskedChoices))
			for i, choice := range msg.Round6Response.ReceiversMaskedChoices {
				compressedReceiversMaskedChoice[i] = choice
			}
			challenges, err := s.bob.Round7DkgRound3Ot(compressedReceiversMaskedChoice)
			if err != nil {
				return err
			}
			challengesBytes := make([][]byte, len(challenges))
			for i, challenge := range challenges {
				challengesBytes[i] = challenge[:]
			}
			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_Round7Response{
					Round7Response: &pb.Round7Response{
						OtChallenges: challengesBytes,
					},
				},
			}); err != nil {
				return err
			}
			// Round 9
		case *pb.DkgMessage_Round8Response:
			fmt.Println("라운드9")

			challengeResponses := make([]simplest.OtChallengeResponse, len(msg.Round8Response.OtChallengeResponses))
			for i, response := range msg.Round8Response.OtChallengeResponses {
				copy(challengeResponses[i][:], response)
			}
			challengeOpenings, err := s.bob.Round9DkgRound5Ot(challengeResponses)
			if err != nil {
				return err
			}

			// 변환 작업
			challengeOpeningsBytes := make([][]byte, len(challengeOpenings))
			for i, opening := range challengeOpenings {
				// opening은 [2][32]byte 타입이므로, 이를 []byte로 변환
				challengeOpeningsBytes[i] = make([]byte, 2*32)
				copy(challengeOpeningsBytes[i][0:32], opening[0][:])
				copy(challengeOpeningsBytes[i][32:], opening[1][:])
			}

			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_Round9Response{
					Round9Response: &pb.Round9Response{
						ChallengeOpenings: challengeOpeningsBytes,
					},
				},
			}); err != nil {
				return err
			}
		case *pb.DkgMessage_Round10Response:
			fmt.Println("라운드끝")

			aliceSecretKeyShare := msg.Round10Response.AliceSecretKeyShare
			curve := curves.K256()
			aliceSecretKey, err := curve.Scalar.SetBytes(aliceSecretKeyShare)
			if err != nil {
				return err
			}
			pkA := curve.ScalarBaseMult(aliceSecretKey)
			computedPublicKeyA := pkA.Mul(s.bob.Output().SecretKeyShare)
			publicKeyBytes := computedPublicKeyA.ToAffineUncompressed()
			publicKeyUnmarshal, err := crypto.UnmarshalPubkey(publicKeyBytes)
			if err != nil {
				log.Fatalf("Failed to unmarshal public key: %v", err)
			}
			address := crypto.PubkeyToAddress(*publicKeyUnmarshal)

			// DKG 완료
			if err := stream.Send(&pb.DkgMessage{
				Msg: &pb.DkgMessage_KeyGenResponse{
					KeyGenResponse: &pb.KeyGenResponse{
						Success: true,
						Address: address.Hex(),
					},
				},
			}); err != nil {
				return err
			}
		}
	}
}
