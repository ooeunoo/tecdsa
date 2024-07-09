package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"tecdsa/pkg/network"
	pb "tecdsa/proto/sign"
	"tecdsa/proto/transaction"

	"google.golang.org/grpc"
)

func SignHandler(w http.ResponseWriter, r *http.Request) {
	var req pb.SignRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		fmt.Printf("Failed to parse request body: %v\n", err)
		return
	}

	networkType, err := network.GetNetworkByID(int(req.Network))
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid network ID: %v", err), http.StatusBadRequest)
		return
	}

	var txMessage *transaction.Transaction
	switch networkType {
	case network.Ethereum, network.Polygon, network.BinanceSmartChain, network.Avalanche:
	case network.Bitcoin, network.Litecoin, network.BitcoinCash:
		if btcTx := req.Transaction.GetBitcoin(); btcTx != nil {
			txMessage = &transaction.Transaction{
				Tx: &transaction.Transaction_Bitcoin{Bitcoin: btcTx},
			}
		} else {
			http.Error(w, "Invalid Bitcoin-like transaction", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Unsupported network for transaction signing", http.StatusBadRequest)
		return
	}

	fmt.Println(txMessage)

	// 채널 생성
	bobChan := make(chan *pb.SignMessage)
	aliceChan := make(chan *pb.SignMessage)
	errorChan := make(chan error)

	// Bob과 Alice 연결 및 스트림 설정
	bobStream, aliceStream, err := setupStreams()
	if err != nil {
		http.Error(w, "Failed to setup streams", http.StatusInternalServerError)
		return
	}

	// Goroutine을 사용하여 Bob과 Alice 간의 메시지 교환
	go receiveMessages(bobStream, bobChan, errorChan)
	go receiveMessages(aliceStream, aliceChan, errorChan)

	// 서명 프로토콜 시작
	if err := aliceStream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_Round1Request{
			Round1Request: &pb.Round1Request{
				Address:   req.Address,
				SecretKey: req.SecretKey,
				// TODO: include tx message
			},
		},
	}); err != nil {
		http.Error(w, "Failed to send initial request to Alice", http.StatusInternalServerError)
		fmt.Printf("Failed to send initial request to Alice: %v\n", err)
		return
	}

	for {
		select {
		case bobResp := <-bobChan:
			if signResp, ok := bobResp.Msg.(*pb.SignMessage_SignResponse); ok {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":   true,
					"signature": signResp.SignResponse.Signature,
				})
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
			http.Error(w, "Error during signing protocol", http.StatusInternalServerError)
			fmt.Printf("Error during signing protocol: %v\n", err)
			return
		}
	}
}

func setupStreams() (pb.SignService_SignClient, pb.SignService_SignClient, error) {
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

func receiveMessages(stream pb.SignService_SignClient, msgChan chan<- *pb.SignMessage, errChan chan<- error) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			errChan <- err
			return
		}
		msgChan <- resp
	}
}
