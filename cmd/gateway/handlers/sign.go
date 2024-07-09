package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	pb "tecdsa/proto/sign"

	"google.golang.org/grpc"
)

func SignHandler(w http.ResponseWriter, r *http.Request) {
	// startTime := time.Now()

	// 요청 바디
	var req pb.Round1Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		fmt.Printf("Failed to parse request body: %v\n", err)
		return
	}

	// Validate request data
	// if req.Address == "" || len(req.SecretKey) == 0 {
	// 	http.Error(w, "Missing required fields", http.StatusBadRequest)
	// 	return
	// }

	// 채널 생성
	bobChan := make(chan *pb.SignMessage)
	aliceChan := make(chan *pb.SignMessage)
	errorChan := make(chan error)

	// Bob과 연결
	bobConn, err := grpc.Dial("bob:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to Bob", http.StatusInternalServerError)
		fmt.Printf("Failed to connect to Bob: %v\n", err)
		return
	}
	defer bobConn.Close()
	bobClient := pb.NewSignServiceClient(bobConn)

	// Alice와 연결
	aliceConn, err := grpc.Dial("alice:50052", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to Alice", http.StatusInternalServerError)
		fmt.Printf("Failed to connect to Alice: %v\n", err)
		return
	}
	defer aliceConn.Close()
	aliceClient := pb.NewSignServiceClient(aliceConn)

	// Bob의 KeyGen 스트림 시작
	bobStream, err := bobClient.Sign(context.Background())
	if err != nil {
		http.Error(w, "Failed to create Bob stream", http.StatusInternalServerError)
		fmt.Printf("Failed to create Bob stream: %v\n", err)
		return
	}

	// Alice의 KeyGen 스트림 시작
	aliceStream, err := aliceClient.Sign(context.Background())
	if err != nil {
		http.Error(w, "Failed to create Alice stream", http.StatusInternalServerError)
		fmt.Printf("Failed to create Alice stream: %v\n", err)
		return
	}

	// Goroutine을 사용하여 Bob과 Alice 간의 메시지 교환
	go func() {
		for {
			bobResp, err := bobStream.Recv()
			if err != nil {
				errorChan <- err
				return
			}
			bobChan <- bobResp
		}
	}()

	go func() {
		for {
			aliceResp, err := aliceStream.Recv()
			if err != nil {
				errorChan <- err
				return
			}
			aliceChan <- aliceResp
		}
	}()

	// DKG 프로토콜 시작 (서명은 Alice부터 라운드 시작)
	if err := aliceStream.Send(&pb.SignMessage{Msg: &pb.SignMessage_Round1Request{Round1Request: &pb.Round1Request{
		Address:   req.Address,
		SecretKey: req.SecretKey,
	}}}); err != nil {
		http.Error(w, "Failed to send initial request to Bob", http.StatusInternalServerError)
		fmt.Printf("Failed to send initial request to Bob: %v\n", err)
		return
	}

	for {
		select {
		case bobResp := <-bobChan:
			// if keyGenResp, ok := bobResp.Msg.(*pb.SignMessage_SignResponse); ok {
			// 	endTime := time.Now()
			// 	duration := endTime.Sub(startTime)
			// 	fmt.Println("생성된 주소:", keyGenResp.SignResponse.Address)
			// 	json.NewEncoder(w).Encode(map[string]interface{}{
			// 		"success":  true,
			// 		"address":  keyGenResp.KeyGenResponse.Address,
			// 		"duration": duration.Seconds(),
			// 	})
			// 	return
			// }
			if err := aliceStream.Send(bobResp); err != nil {
				http.Error(w, "Failed to send Bob's response to Alice", http.StatusInternalServerError)
				fmt.Printf("Failed to send Bob's response to Alice: %v\n", err)
				return
			}

		case aliceResp := <-aliceChan:
			if err := bobStream.Send(aliceResp); err != nil {
				http.Error(w, "Failed to send Alice's response to Bob", http.StatusInternalServerError)
				fmt.Printf("Failed to send Alice's response to Bob: %v\n", err)
				return
			}

		case err := <-errorChan:
			http.Error(w, "Error during DKG protocol", http.StatusInternalServerError)
			fmt.Printf("Error during DKG protocol: %v\n", err)
			return
		}
	}
}
