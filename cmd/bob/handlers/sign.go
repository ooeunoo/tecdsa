// dkg_handler.go
package handlers

import (
	"io"
	"tecdsa/internal/dkls/sign"
	pb "tecdsa/pkg/api/grpc/sign"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

type SignHandler struct {
	curve *curves.Curve
}

func NewSignHandler() *SignHandler {
	return &SignHandler{
		curve: curves.K256(),
	}
}

func (h *SignHandler) HandleSign(stream pb.SignService_SignServer) error {
	// bob := dkg.NewBob(h.curve)

	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// switch msg := in.Msg.(type) {
		// case *pb.DkgMessage_Round1Request:
		// 	err = h.handleRound1(stream, bob)
		// case *pb.DkgMessage_Round2Response:
		// 	err = h.handleRound3(stream, bob, msg.Round2Response)
		// case *pb.DkgMessage_Round4Response:
		// 	err = h.handleRound5(stream, bob, msg.Round4Response)
		// case *pb.DkgMessage_Round6Response:
		// 	err = h.handleRound7(stream, bob, msg.Round6Response)
		// case *pb.DkgMessage_Round8Response:
		// 	err = h.handleRound9(stream, bob, msg.Round8Response)
		// case *pb.DkgMessage_Round10Response:
		// 	err = h.handleFinalRound(stream, bob)
		// default:
		// 	err = fmt.Errorf("unexpected message type")
		// }

		// if err != nil {
		// 	return err
		// }
	}
}

func (h *SignHandler) handleRound2(stream pb.SignService_SignServer, bob *sign.Bob, msg *pb.Round1Response) error {

	return stream.Send(&pb.SignMessage{})
}

func (h *SignHandler) handleRound4(stream pb.SignService_SignServer, bob *sign.Bob, msg *pb.Round3Response) error {

	return stream.Send(&pb.SignMessage{})
}
