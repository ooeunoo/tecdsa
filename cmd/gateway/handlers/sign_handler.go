package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"tecdsa/cmd/gateway/config"
	"tecdsa/pkg/database/repository"
	"tecdsa/pkg/response"
	"tecdsa/pkg/service"
	"tecdsa/pkg/utils"
	pb "tecdsa/proto/sign"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type SignRequest struct {
	Address   string `json:"address"`
	TxOrigin  string `json:"tx_origin"`
	RequestID string `json:"request_id,omitempty"`
}

type SignResponse struct {
	V         uint64 `json:"v"`
	R         string `json:"r"`
	S         string `json:"s"`
	Duration  int32  `json:"duration"`
	RequestID string `json:"request_id"`
}

type signRequestContext struct {
	startTime        time.Time
	address          string
	txOrigin         string
	clientSecurityID uint32
}

type SignHandler struct {
	clientSecurityRepo repository.ClientSecurityRepository
	config             *config.Config
	networkService     *service.NetworkService
	requestContexts    map[string]*signRequestContext
	mutex              sync.Mutex
}

func NewSignHandler(cfg *config.Config, repo repository.ClientSecurityRepository, networkService *service.NetworkService) *SignHandler {
	return &SignHandler{
		clientSecurityRepo: repo,
		config:             cfg,
		networkService:     networkService,
		requestContexts:    make(map[string]*signRequestContext),
	}
}

func (h *SignHandler) Serve(w http.ResponseWriter, r *http.Request) {
	req, requestID, err := h.parseAndValidateSignRequest(r)
	if err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, err.Error()))
		return
	}

	clientIP := utils.GetClientIP(r)
	clientSecurity, err := h.clientSecurityRepo.FindByIP(clientIP)
	if err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeInternalServerError, response.ErrMsgFailedRetrieveClientSecurity))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	ctx = h.addSignMetadataToContext(ctx, requestID, req, clientSecurity.ID)

	if err := h.storeSignRequestContext(requestID, req, clientSecurity.ID); err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, err.Error()))
		return
	}
	defer h.removeSignRequestContext(requestID)

	bobStream, aliceStream, err := h.setupSignStreams(ctx)
	if err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeInternalServerError, response.ErrMsgFailedSetupStreams))
		return
	}
	defer bobStream.CloseSend()
	defer aliceStream.CloseSend()

	if err := h.performSigning(w, bobStream, aliceStream, requestID); err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeSigning, err.Error()))
	}
}

func (h *SignHandler) parseAndValidateSignRequest(r *http.Request) (SignRequest, string, error) {
	var req SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, "", fmt.Errorf(response.ErrMsgInvalidRequestBody)
	}

	requestID := strings.TrimSpace(req.RequestID)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	if req.Address == "" || req.TxOrigin == "" {
		return req, "", fmt.Errorf(response.ErrMsgInvalidSignRequest)
	}

	return req, requestID, nil
}

func (h *SignHandler) addSignMetadataToContext(ctx context.Context, requestID string, req SignRequest, clientSecurityID uint32) context.Context {
	md := metadata.New(map[string]string{
		"request_id":         requestID,
		"address":            req.Address,
		"tx_origin":          req.TxOrigin,
		"client_security_id": fmt.Sprintf("%d", clientSecurityID),
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func (h *SignHandler) storeSignRequestContext(requestID string, req SignRequest, clientSecurityID uint32) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.requestContexts[requestID]; exists {
		return fmt.Errorf(response.ErrMsgDuplicateRequestID)
	}

	h.requestContexts[requestID] = &signRequestContext{
		startTime:        time.Now(),
		address:          req.Address,
		txOrigin:         req.TxOrigin,
		clientSecurityID: clientSecurityID,
	}

	return nil
}

func (h *SignHandler) removeSignRequestContext(requestID string) {
	h.mutex.Lock()
	delete(h.requestContexts, requestID)
	h.mutex.Unlock()
}

func (h *SignHandler) performSigning(w http.ResponseWriter, bobStream, aliceStream pb.SignService_SignClient, requestID string) error {
	bobChan := make(chan *pb.SignMessage)
	aliceChan := make(chan *pb.SignMessage)
	errorChan := make(chan error)

	go h.receiveSignMessages(bobStream, bobChan, errorChan)
	go h.receiveSignMessages(aliceStream, aliceChan, errorChan)

	if err := h.startSignProtocol(aliceStream, requestID); err != nil {
		return fmt.Errorf(response.ErrMsgFailedStartSigning)
	}

	return h.handleSignMessages(w, bobStream, aliceStream, bobChan, aliceChan, errorChan, requestID)
}

func (h *SignHandler) startSignProtocol(aliceStream pb.SignService_SignClient, requestID string) error {
	return aliceStream.Send(&pb.SignMessage{
		Msg: &pb.SignMessage_SignGatewayTo1Output{
			SignGatewayTo1Output: &pb.SignGatewayTo1Output{},
		},
	})
}

func (h *SignHandler) handleSignMessages(w http.ResponseWriter, bobStream, aliceStream pb.SignService_SignClient, bobChan, aliceChan <-chan *pb.SignMessage, errorChan <-chan error, requestID string) error {
	for {
		select {
		case bobResp := <-bobChan:
			if signResp, ok := bobResp.Msg.(*pb.SignMessage_SignRound4ToGatewayOutput); ok {
				return h.handleFinalSignResponse(w, signResp, requestID)
			}
			if err := aliceStream.Send(bobResp); err != nil {
				return fmt.Errorf(response.ErrMsgFailedDuringSigning)
			}
		case aliceResp := <-aliceChan:
			if err := bobStream.Send(aliceResp); err != nil {
				return fmt.Errorf(response.ErrMsgFailedDuringSigning)
			}
		case <-errorChan:
			return fmt.Errorf(response.ErrMsgFailedDuringSigning)
		}
	}
}

func (h *SignHandler) handleFinalSignResponse(w http.ResponseWriter, signResp *pb.SignMessage_SignRound4ToGatewayOutput, requestID string) error {
	h.mutex.Lock()
	reqCtx, exists := h.requestContexts[requestID]
	h.mutex.Unlock()

	if !exists {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, response.ErrMsgInvalidRequestID))
		return fmt.Errorf(response.ErrMsgInvalidRequestID)
	}

	duration := time.Since(reqCtx.startTime)

	signResponse := SignResponse{
		V:         signResp.SignRound4ToGatewayOutput.V,
		R:         base64.StdEncoding.EncodeToString(signResp.SignRound4ToGatewayOutput.R),
		S:         base64.StdEncoding.EncodeToString(signResp.SignRound4ToGatewayOutput.S),
		Duration:  int32(duration.Milliseconds()),
		RequestID: requestID,
	}

	response.SendResponse(w, response.NewSuccessResponse(http.StatusOK, signResponse))
	return nil
}

func (h *SignHandler) setupSignStreams(ctx context.Context) (pb.SignService_SignClient, pb.SignService_SignClient, error) {
	bobStream, err := h.setupSignStream(ctx, h.config.BobGRPCAddress)
	if err != nil {
		return nil, nil, fmt.Errorf(response.ErrMsgFailedSetupStreams)
	}

	aliceStream, err := h.setupSignStream(ctx, h.config.AliceGRPCAddress)
	if err != nil {
		return nil, nil, fmt.Errorf(response.ErrMsgFailedSetupStreams)
	}

	return bobStream, aliceStream, nil
}

func (h *SignHandler) setupSignStream(ctx context.Context, address string) (pb.SignService_SignClient, error) {
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf(response.ErrMsgFailedConnectGRPC)
	}
	client := pb.NewSignServiceClient(conn)
	return client.Sign(ctx)
}

func (h *SignHandler) receiveSignMessages(stream pb.SignService_SignClient, msgChan chan<- *pb.SignMessage, errChan chan<- error) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			errChan <- err
			return
		}
		msgChan <- resp
	}
}
