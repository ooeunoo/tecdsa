package handlers

import (
	"net/http"
	"tecdsa/pkg/response"
	"tecdsa/pkg/service"
)

type NetworkInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type GetAllNetworksHandler struct {
	networkService *service.NetworkService
}

func NewGetAllNetworksHandler(networkService *service.NetworkService) *GetAllNetworksHandler {
	return &GetAllNetworksHandler{
		networkService: networkService,
	}
}

func (h *GetAllNetworksHandler) Serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	networks := h.networkService.GetAllNetworks()

	var networkInfos []NetworkInfo
	for _, net := range networks {
		networkInfos = append(networkInfos, NetworkInfo{
			ID:   int(net),
			Name: net.String(),
		})
	}

	response.SendResponse(w, response.NewSuccessResponse(http.StatusOK, map[string]interface{}{
		"networks": networkInfos,
	}))
}
