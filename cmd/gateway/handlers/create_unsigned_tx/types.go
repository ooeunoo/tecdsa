package handlers

import "tecdsa/pkg/service"

type CreateUnsignedTxHandler struct {
	networkService *service.NetworkService
}