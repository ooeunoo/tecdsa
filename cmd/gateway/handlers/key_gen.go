package handlers

import (
	"context"
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
	pb "tecdsa/proto/keygen"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type KeyGenRequest struct {
	RequestID string `json:"request_id,omitempty"`
	Network   int32  `json:"network"`
}

type KeyGenResponse struct {
	RequestID string `json:"request_id"`
	Address   string `json:"address"`
	Publickey string `json:"public_key"`
	Duration  int32  `json:"duration"`
}

type requestContext struct {
	startTime        time.Time
	network          int32
	clientSecurityID uint32
}

type KeyGenHandler struct {
	clientSecurityRepo repository.ClientSecurityRepository
	config             *config.Config
	networkService     *service.NetworkService
	requestContexts    map[string]*requestContext
	mutex              sync.Mutex
}

func NewKeyGenHandler(cfg *config.Config, repo repository.ClientSecurityRepository, networkService *service.NetworkService) *KeyGenHandler {
	return &KeyGenHandler{
		clientSecurityRepo: repo,
		config:             cfg,
		networkService:     networkService,
		requestContexts:    make(map[string]*requestContext),
	}
}

func (h *KeyGenHandler) Serve(w http.ResponseWriter, r *http.Request) {
	req, requestID, err := h.parseAndValidateRequest(r)
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

	ctx = h.addMetadataToContext(ctx, requestID, req, uint32(clientSecurity.ID))

	if err := h.storeRequestContext(requestID, req, uint32(clientSecurity.ID)); err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, err.Error()))
		return
	}
	defer h.removeRequestContext(requestID)

	bobStream, aliceStream, err := h.setupKeygenStreams(ctx)
	if err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeInternalServerError, response.ErrMsgFailedSetupStreams))
		return
	}
	defer bobStream.CloseSend()
	defer aliceStream.CloseSend()

	if err := h.performKeyGeneration(w, bobStream, aliceStream, requestID); err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeKeyGeneration, err.Error()))
	}
}

func (h *KeyGenHandler) parseAndValidateRequest(r *http.Request) (KeyGenRequest, string, error) {
	var req KeyGenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, "", fmt.Errorf(response.ErrMsgInvalidRequestBody)
	}

	requestID := strings.TrimSpace(req.RequestID)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	if _, err := h.networkService.GetNetworkByID(req.Network); err != nil {
		return req, "", fmt.Errorf(response.ErrMsgUnsupportedNetwork)
	}

	return req, requestID, nil
}

func (h *KeyGenHandler) addMetadataToContext(ctx context.Context, requestID string, req KeyGenRequest, clientSecurityID uint32) context.Context {
	md := metadata.New(map[string]string{
		"request_id":         requestID,
		"network":            fmt.Sprintf("%d", req.Network),
		"client_security_id": fmt.Sprintf("%d", clientSecurityID),
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func (h *KeyGenHandler) storeRequestContext(requestID string, req KeyGenRequest, clientSecurityID uint32) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.requestContexts[requestID]; exists {
		return fmt.Errorf(response.ErrMsgDuplicateRequestID)
	}

	h.requestContexts[requestID] = &requestContext{
		startTime:        time.Now(),
		network:          req.Network,
		clientSecurityID: clientSecurityID,
	}

	return nil
}

func (h *KeyGenHandler) removeRequestContext(requestID string) {
	h.mutex.Lock()
	delete(h.requestContexts, requestID)
	h.mutex.Unlock()
}

func (h *KeyGenHandler) performKeyGeneration(w http.ResponseWriter, bobStream, aliceStream pb.KeygenService_KeyGenClient, requestID string) error {
	bobChan := make(chan *pb.KeygenMessage)
	aliceChan := make(chan *pb.KeygenMessage)
	errorChan := make(chan error)

	go h.receiveKeygenMessages(bobStream, bobChan, errorChan)
	go h.receiveKeygenMessages(aliceStream, aliceChan, errorChan)

	if err := h.startDKGProtocol(bobStream, requestID); err != nil {
		return fmt.Errorf(response.ErrMsgFailedStartKeyGeneration)
	}

	return h.handleKeyGenMessages(w, bobStream, aliceStream, bobChan, aliceChan, errorChan, requestID)
}

func (h *KeyGenHandler) startDKGProtocol(bobStream pb.KeygenService_KeyGenClient, requestID string) error {
	return bobStream.Send(&pb.KeygenMessage{Msg: &pb.KeygenMessage_KeyGenGatewayTo1Output{
		KeyGenGatewayTo1Output: &pb.KeyGenGatewayTo1Output{},
	}})
}

func (h *KeyGenHandler) handleKeyGenMessages(w http.ResponseWriter, bobStream, aliceStream pb.KeygenService_KeyGenClient, bobChan, aliceChan <-chan *pb.KeygenMessage, errorChan <-chan error, requestID string) error {
	for {
		select {
		case bobResp := <-bobChan:
			if res, ok := bobResp.Msg.(*pb.KeygenMessage_KeyGenRound11ToGatewayOutput); ok {
				return h.handleFinalResponse(w, res, requestID)
			}
			if err := aliceStream.Send(bobResp); err != nil {
				return fmt.Errorf(response.ErrMsgFailedDuringKeyGeneration)
			}
		case aliceResp := <-aliceChan:
			if err := bobStream.Send(aliceResp); err != nil {
				return fmt.Errorf(response.ErrMsgFailedDuringKeyGeneration)
			}
		case <-errorChan:
			return fmt.Errorf(response.ErrMsgFailedDuringKeyGeneration)
		}
	}
}

func (h *KeyGenHandler) handleFinalResponse(w http.ResponseWriter, res *pb.KeygenMessage_KeyGenRound11ToGatewayOutput, requestID string) error {
	h.mutex.Lock()
	reqCtx, exists := h.requestContexts[requestID]
	h.mutex.Unlock()

	if !exists {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, response.ErrMsgInvalidRequestID))
		return fmt.Errorf(response.ErrMsgInvalidRequestID)
	}

	duration := time.Since(reqCtx.startTime)

	keyGenResponse := KeyGenResponse{
		RequestID: requestID,
		Address:   res.KeyGenRound11ToGatewayOutput.Address,
		Publickey: res.KeyGenRound11ToGatewayOutput.PublicKey,
		Duration:  int32(duration.Milliseconds()),
	}

	response.SendResponse(w, response.NewSuccessResponse(http.StatusOK, keyGenResponse))
	return nil
}

func (h *KeyGenHandler) setupKeygenStreams(ctx context.Context) (pb.KeygenService_KeyGenClient, pb.KeygenService_KeyGenClient, error) {
	bobStream, err := h.setupKeygenStream(ctx, h.config.BobGRPCAddress)
	if err != nil {
		return nil, nil, fmt.Errorf(response.ErrMsgFailedSetupStreams)
	}

	aliceStream, err := h.setupKeygenStream(ctx, h.config.AliceGRPCAddress)
	if err != nil {
		return nil, nil, fmt.Errorf(response.ErrMsgFailedSetupStreams)
	}

	return bobStream, aliceStream, nil
}

func (h *KeyGenHandler) setupKeygenStream(ctx context.Context, address string) (pb.KeygenService_KeyGenClient, error) {
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf(response.ErrMsgFailedConnectGRPC)
	}
	client := pb.NewKeygenServiceClient(conn)
	return client.KeyGen(ctx)
}

func (h *KeyGenHandler) receiveKeygenMessages(stream pb.KeygenService_KeyGenClient, msgChan chan<- *pb.KeygenMessage, errChan chan<- error) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			errChan <- err
			return
		}
		msgChan <- resp
	}
}
