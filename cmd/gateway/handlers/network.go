package handlers

import (
	"net/http"
	"tecdsa/pkg/network"
	"tecdsa/pkg/response"
)

type NetworkInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func GetAllNetworksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var networks []NetworkInfo

	// 모든 네트워크를 순회합니다.
	for id := 1; id < len(network.Networks)+1; id++ {
		network, err := network.GetNetworkByID(int32(id))
		if err == nil {
			networks = append(networks, NetworkInfo{
				ID:   id,
				Name: network.String(),
			})
		}
	}

	response.SendResponse(w, response.NewSuccessResponse(http.StatusOK, map[string]interface{}{
		"networks": networks,
	}))
}
