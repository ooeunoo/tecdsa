package handlers

import (
	"net/http"

	"tecdsa/internal/encoding/dkg"
	pb "tecdsa/pkg/api/grpc/dkg"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DkgHandler struct {
	aliceClient pb.DkgServiceClient
	bobClient   pb.DkgServiceClient
}

func NewDkgHandler(aliceClient, bobClient pb.DkgServiceClient) *DkgHandler {
	return &DkgHandler{
		aliceClient: aliceClient,
		bobClient:   bobClient,
	}
}

func (h *DkgHandler) HandleDkg(c *gin.Context) {
	sessionID := uuid.New().String()

	// Start DKG process with Bob
	bobResp, err := h.bobClient.StartDkg(c, &pb.DkgRequest{SessionId: sessionID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start DKG with Bob"})
		return
	}

	// Continue DKG process with Alice
	aliceResp, err := h.aliceClient.ContinueDkg(c, &pb.DkgContinueRequest{
		SessionId: sessionID,
		Data:      bobResp.Data,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to continue DKG with Alice"})
		return
	}

	// Finish DKG process with Bob
	finalResp, err := h.bobClient.FinishDkg(c, &pb.DkgFinishRequest{
		SessionId: sessionID,
		Data:      aliceResp.Data,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finish DKG with Bob"})
		return
	}

	// Decode and combine results
	result, err := dkg.CombineResults(aliceResp, finalResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to combine results"})
		return
	}

	c.JSON(http.StatusOK, result)
}
