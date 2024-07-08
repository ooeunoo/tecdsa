package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	pb "tecdsa/pkg/api/grpc/dkg"

	"google.golang.org/grpc"
)

func KeyGenHandler(w http.ResponseWriter, r *http.Request) {
	// Bob과 연결
	bobConn, err := grpc.Dial("bob:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to Bob", http.StatusInternalServerError)
		return
	}
	defer bobConn.Close()
	bobClient := pb.NewDkgServiceClient(bobConn)

	// Alice와 연결
	aliceConn, err := grpc.Dial("alice:50052", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to Alice", http.StatusInternalServerError)
		return
	}
	defer aliceConn.Close()
	aliceClient := pb.NewDkgServiceClient(aliceConn)

	// Bob의 KeyGen 스트림 시작
	bobStream, err := bobClient.KeyGen(context.Background())
	if err != nil {
		http.Error(w, "Failed to create Bob stream", http.StatusInternalServerError)
		return
	}

	// Alice의 KeyGen 스트림 시작
	aliceStream, err := aliceClient.KeyGen(context.Background())
	if err != nil {
		http.Error(w, "Failed to create Alice stream", http.StatusInternalServerError)
		return
	}

	// DKG 프로토콜 시작
	if err := bobStream.Send(&pb.DkgMessage{Msg: &pb.DkgMessage_Round1Request{Round1Request: &pb.Round1Request{}}}); err != nil {
		http.Error(w, "Failed to send initial request to Bob", http.StatusInternalServerError)
		return
	}

	// Bob과 Alice 사이의 메시지 교환
	for {
		// Bob으로부터 응답 받기
		bobResp, err := bobStream.Recv()

		if err != nil {
			http.Error(w, "Failed to receive response from Bob", http.StatusInternalServerError)
			return
		}
		// DKG 프로토콜 완료 확인
		if keyGenResp, ok := bobResp.Msg.(*pb.DkgMessage_KeyGenResponse); ok {
			fmt.Println("Received KeyGenResponse with address:", keyGenResp.KeyGenResponse.Address)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"address": keyGenResp.KeyGenResponse.Address,
			})
			return
		}

		// Bob의 응답을 Alice에게 전송
		if err := aliceStream.Send(bobResp); err != nil {
			http.Error(w, "Failed to send Bob's response to Alice", http.StatusInternalServerError)
			return
		}

		// Alice로부터 응답 받기
		aliceResp, err := aliceStream.Recv()
		if err != nil {
			http.Error(w, "Failed to receive response from Alice", http.StatusInternalServerError)
			return
		}

		// Alice의 응답을 Bob에게 전송
		if err := bobStream.Send(aliceResp); err != nil {
			http.Error(w, "Failed to send Alice's response to Bob", http.StatusInternalServerError)
			return
		}

	}
}
