package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	pb "tecdsa/pkg/api/grpc/dkg"

	"google.golang.org/grpc"
)

func DKGHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RandomNumber string `json:"random_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := grpc.Dial("bob:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to Bob", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := pb.NewDKGServiceClient(conn)
	resp, err := client.ProcessDKG(context.Background(), &pb.DKGRequest{RandomNumber: req.RandomNumber})
	if err != nil {
		http.Error(w, "Error calling Bob's ProcessDKG", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"result": resp.Result})
}
