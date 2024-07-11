package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tecdsa/pkg/response"
	pb "tecdsa/proto/sign"

	"google.golang.org/grpc"
)

type SignRequestDTO struct {
	Address   string `json:"address"`
	SecretKey string `json:"secretKey"`
	TxOrigin  string `json:"txOrigin"`
}

func SignHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req pb.SignRequestMessage
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		fmt.Printf("Failed to parse request body: %v\n", err)
		return
	}

	secretKeyBytes, _ := base64.StdEncoding.DecodeString(req.SecretKey)
	txOriginBytes, _ := base64.StdEncoding.DecodeString(req.TxOrigin)

	// 채널 생성
	bobChan := make(chan *pb.SignMessage)
	aliceChan := make(chan *pb.SignMessage)
	errorChan := make(chan error)

	// Bob과 Alice 연결 및 스트림 설정
	bobStream, aliceStream, err := setupSignStreams()
	if err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeSigning))
		return
	}

	// Goroutine을 사용하여 Bob과 Alice 간의 메시지 교환
	go receiveSignMessages(bobStream, bobChan, errorChan)
	go receiveSignMessages(aliceStream, aliceChan, errorChan)

	// 서명 프로토콜 시작
	if err := aliceStream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_SignRequestTo1Output{
			SignRequestTo1Output: &pb.SignRequestTo1Output{
				Address:   req.Address,
				SecretKey: secretKeyBytes,
				TxOrigin:  txOriginBytes,
			},
		},
	}); err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeSigning))
		fmt.Printf("Failed to send initial request to Alice: %v\n", err)
		return
	}

	for {
		select {
		case bobResp := <-bobChan:
			if signResp, ok := bobResp.Msg.(*pb.SignMessage_SignRound4ToResponseOutput); ok {
				endTime := time.Now()
				duration := endTime.Sub(startTime)

				signResponse := &pb.SignResponse{
					V:        signResp.SignRound4ToResponseOutput.V,
					R:        base64.StdEncoding.EncodeToString(signResp.SignRound4ToResponseOutput.R),
					S:        base64.StdEncoding.EncodeToString(signResp.SignRound4ToResponseOutput.S),
					Duration: int32(duration.Milliseconds()),
				}

				response.SendResponse(w, response.NewSuccessResponse(http.StatusOK, signResponse))
				return
			}
			if err := aliceStream.Send(bobResp); err != nil {
				response.SendResponse(w, response.NewErrorResponse(response.ErrCodeSigning))
				fmt.Printf("Failed to send Bob's response to Alice: %v\n", err)
				return
			}

		case aliceResp := <-aliceChan:
			if err := bobStream.Send(aliceResp); err != nil {
				response.SendResponse(w, response.NewErrorResponse(response.ErrCodeSigning))
				fmt.Printf("Failed to send Alice's response to Bob: %v\n", err)
				return
			}

		case err := <-errorChan:
			response.SendResponse(w, response.NewErrorResponse(response.ErrCodeSigning))
			fmt.Printf("Error during signing protocol: %v\n", err)
			return
		}
	}
}

func receiveSignMessages(stream pb.SignService_SignClient, msgChan chan<- *pb.SignMessage, errChan chan<- error) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			errChan <- err
			return
		}
		msgChan <- resp
	}
}

func setupSignStreams() (pb.SignService_SignClient, pb.SignService_SignClient, error) {
	// Bob과 연결
	bobConn, err := grpc.Dial("bob:50051", grpc.WithInsecure())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Bob: %v", err)
	}
	bobClient := pb.NewSignServiceClient(bobConn)
	bobStream, err := bobClient.Sign(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Bob stream: %v", err)
	}

	// Alice와 연결
	aliceConn, err := grpc.Dial("alice:50052", grpc.WithInsecure())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Alice: %v", err)
	}
	aliceClient := pb.NewSignServiceClient(aliceConn)
	aliceStream, err := aliceClient.Sign(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Alice stream: %v", err)
	}

	return bobStream, aliceStream, nil
}
