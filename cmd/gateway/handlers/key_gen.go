package handlers

import (
	"context"
	"fmt"
	"net/http"
	pb "tecdsa/proto/keygen"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func KeyGenHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 채널 생성
	bobChan := make(chan *pb.KeygenMessage)
	aliceChan := make(chan *pb.KeygenMessage)
	errorChan := make(chan error)

	// Bob과 연결
	bobConn, err := grpc.Dial("bob:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to Bob", http.StatusInternalServerError)
		fmt.Printf("Failed to connect to Bob: %v\n", err)
		return
	}
	defer bobConn.Close()
	bobClient := pb.NewKeygenServiceClient(bobConn)

	// Alice와 연결
	aliceConn, err := grpc.Dial("alice:50052", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to Alice", http.StatusInternalServerError)
		fmt.Printf("Failed to connect to Alice: %v\n", err)
		return
	}
	defer aliceConn.Close()
	aliceClient := pb.NewKeygenServiceClient(aliceConn)

	// Bob의 KeyGen 스트림 시작
	bobStream, err := bobClient.KeyGen(context.Background())
	if err != nil {
		http.Error(w, "Failed to create Bob stream", http.StatusInternalServerError)
		fmt.Printf("Failed to create Bob stream: %v\n", err)
		return
	}

	// Alice의 KeyGen 스트림 시작
	aliceStream, err := aliceClient.KeyGen(context.Background())
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
				fmt.Println("bob")

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
				fmt.Println("alice")
				errorChan <- err
				return
			}
			aliceChan <- aliceResp
		}
	}()

	// DKG 프로토콜 시작
	if err := bobStream.Send(&pb.KeygenMessage{Msg: &pb.KeygenMessage_KeyGenRequestTo1Output{KeyGenRequestTo1Output: &pb.KeyGenRequestTo1Output{}}}); err != nil {
		http.Error(w, "Failed to send initial request to Bob", http.StatusInternalServerError)
		fmt.Printf("Failed to send initial request to Bob: %v\n", err)
		return
	}

	for {
		select {
		case bobResp := <-bobChan:
			if res, ok := bobResp.Msg.(*pb.KeygenMessage_KeyGenRound11ToResponseOutput); ok {
				endTime := time.Now()
				duration := endTime.Sub(startTime)

				// SignResponse 메시지 생성
				response := &pb.KeyGenResponse{
					Success:   true,
					Address:   res.KeyGenRound11ToResponseOutput.Address,
					SecretKey: res.KeyGenRound11ToResponseOutput.SecretKey,
					Duration:  int32(duration.Milliseconds()),
				}

				marshaler := protojson.MarshalOptions{
					EmitUnpopulated: true,
					UseProtoNames:   true,
				}

				// SignResponse를 JSON으로 변환
				jsonBytes, err := marshaler.Marshal(response)
				if err != nil {
					http.Error(w, "Failed to marshal response to JSON", http.StatusInternalServerError)
					return
				}

				// Content-Type 설정
				w.Header().Set("Content-Type", "application/json")

				// JSON 응답 전송
				w.Write(jsonBytes)
				return

			}
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
