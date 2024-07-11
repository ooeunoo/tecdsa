package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"tecdsa/pkg/response"
	pb "tecdsa/proto/keygen"
	"time"

	"google.golang.org/grpc"
)

func KeyGenHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req pb.KeyGenRequestMessage
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.SendResponse(w, response.NewErrorResponse(http.StatusBadRequest, "Invalid request body"))
		fmt.Printf("Failed to parse request body: %v\n", err)
		return
	}

	// 채널 생성
	bobChan := make(chan *pb.KeygenMessage)
	aliceChan := make(chan *pb.KeygenMessage)
	errorChan := make(chan error)

	// Bob과 Alice 연결 및 스트림 설정

	bobStream, aliceStream, err := setupKeygenStreams()
	if err != nil {
		http.Error(w, "Failed to setup streams", http.StatusInternalServerError)
		return
	}

	// Goroutine을 사용하여 Bob과 Alice 간의 메시지 교환
	go receiveKeygenMessages(bobStream, bobChan, errorChan)
	go receiveKeygenMessages(aliceStream, aliceChan, errorChan)

	// DKG 프로토콜 시작
	if err := bobStream.Send(&pb.KeygenMessage{Msg: &pb.KeygenMessage_KeyGenRequestTo1Output{KeyGenRequestTo1Output: &pb.KeyGenRequestTo1Output{
		Network: req.Network,
	}}}); err != nil {
		response.SendResponse(w, response.NewErrorResponse(http.StatusInternalServerError, "Error during DKG protocol"))
		fmt.Printf("Failed to send initial request to Bob: %v\n", err)
		return
	}

	for {
		select {
		case bobResp := <-bobChan:
			if res, ok := bobResp.Msg.(*pb.KeygenMessage_KeyGenRound11ToResponseOutput); ok {
				endTime := time.Now()
				duration := endTime.Sub(startTime)

				keyGenResponse := &pb.KeyGenResponse{
					Address:   res.KeyGenRound11ToResponseOutput.Address,
					SecretKey: base64.StdEncoding.EncodeToString(res.KeyGenRound11ToResponseOutput.SecretKey),
					Duration:  int32(duration.Milliseconds()),
				}
				fmt.Println("keyGenResponse:", keyGenResponse)
				response.SendResponse(w, response.NewSuccessResponse(http.StatusOK, keyGenResponse))
				return
			}
			if err := aliceStream.Send(bobResp); err != nil {
				response.SendResponse(w, response.NewErrorResponse(http.StatusInternalServerError, "Error during DKG protocol"))
				fmt.Printf("Failed to send Bob's response to Alice: %v\n", err)
				return
			}

		case aliceResp := <-aliceChan:
			if err := bobStream.Send(aliceResp); err != nil {
				response.SendResponse(w, response.NewErrorResponse(http.StatusInternalServerError, "Error during DKG protocol"))
				fmt.Printf("Failed to send Alice's response to Bob: %v\n", err)
				return
			}

		case err := <-errorChan:
			response.SendResponse(w, response.NewErrorResponse(http.StatusInternalServerError, "Error during DKG protocol"))
			fmt.Printf("Error during DKG protocol: %v\n", err)
			return
		}
	}
}

func receiveKeygenMessages(stream pb.KeygenService_KeyGenClient, msgChan chan<- *pb.KeygenMessage, errChan chan<- error) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			errChan <- err
			return
		}
		msgChan <- resp
	}
}

func setupKeygenStreams() (pb.KeygenService_KeyGenClient, pb.KeygenService_KeyGenClient, error) {
	// Bob과 연결
	bobConn, err := grpc.Dial("bob:50051", grpc.WithInsecure())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Bob: %v", err)
	}
	bobClient := pb.NewKeygenServiceClient(bobConn)
	bobStream, err := bobClient.KeyGen(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Bob stream: %v", err)
	}

	// Alice와 연결
	aliceConn, err := grpc.Dial("alice:50052", grpc.WithInsecure())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Alice: %v", err)
	}
	aliceClient := pb.NewKeygenServiceClient(aliceConn)
	aliceStream, err := aliceClient.KeyGen(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Alice stream: %v", err)
	}

	return bobStream, aliceStream, nil
}
