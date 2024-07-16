package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"tecdsa/pkg/network"
	"tecdsa/pkg/response"
	"tecdsa/pkg/service"
)

func NewCreateUnsignedTxHandler(networkService *service.NetworkService) *CreateUnsignedTxHandler {
	return &CreateUnsignedTxHandler{
		networkService: networkService,
	}
}

func (h *CreateUnsignedTxHandler) Serve(w http.ResponseWriter, r *http.Request) {
	networkID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/create_unsigned_tx/"))
	if err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "Invalid network ID"))
		return
	}

	networkType, err := h.networkService.GetNetworkByID(int32(networkID))
	if err != nil {
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "Unsupported network ID"))
		return
	}

	fmt.Println("networkType", networkType)
	var txRequest interface{}
	switch networkType {
	case network.Bitcoin, network.BitcoinTestNet:
		var btcReq network.BitcoinTxRequest
		if err := json.NewDecoder(r.Body).Decode(&btcReq); err != nil {
			response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "Invalid request body"))
			return
		}

		// 필수 필드 유효성 검사
		if btcReq.From == "" {
			response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "'from' is required"))
			return
		}
		if btcReq.To == "" {
			response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "'to' is required"))
			return
		}
		if btcReq.Amount == "" {
			response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "'amount' is required"))
			return
		}
		fmt.Print(btcReq)

		txRequest = btcReq
	case network.Ethereum, network.Ethereum_Sepolia:
		var ethReq network.EthereumTxRequest
		if err := json.NewDecoder(r.Body).Decode(&ethReq); err != nil {
			response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "Invalid request body"))
			return
		}

		// 필수 필드 유효성 검사
		if ethReq.From == "" {
			response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "'from' is required"))
			return
		}
		if ethReq.To == "" {
			response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "'to' is required"))
			return
		}
		if ethReq.Amount == "" {
			response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "'amount' is required"))
			return
		}

		txRequest = ethReq
	default:
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeBadRequest, "Unsupported network type"))
		return
	}

	fmt.Println("herer")
	unsignedTx, err := h.networkService.CreateUnsignedTransaction(networkType, txRequest)
	if err != nil {
		fmt.Printf("Error creating unsigned transaction: %v\n", err)
		response.SendResponse(w, response.NewErrorResponse(response.ErrCodeInternalServerError, fmt.Sprintf("Failed to create unsigned transaction: %v", err)))
		return
	}

	response.SendResponse(w, response.NewSuccessResponse(http.StatusOK, unsignedTx))
}
