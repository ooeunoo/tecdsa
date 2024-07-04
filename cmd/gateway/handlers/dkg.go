package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

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

	originalNumber, err := strconv.Atoi(req.RandomNumber)
	if err != nil {
		http.Error(w, "Invalid random number", http.StatusBadRequest)
		return
	}

	conn, err := grpc.Dial("bob:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to Bob", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := pb.NewDKGServiceClient(conn)
	resp, err := client.ProcessDKG(r.Context(), &pb.DKGRequest{RandomNumber: req.RandomNumber})
	if err != nil {
		http.Error(w, "Error calling Bob's ProcessDKG", http.StatusInternalServerError)
		return
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(resp.Result)
	if err != nil {
		http.Error(w, "Error decoding result", http.StatusInternalServerError)
		return
	}

	decodedResult, err := strconv.Atoi(string(decodedBytes))
	if err != nil {
		http.Error(w, "Error converting result to integer", http.StatusInternalServerError)
		return
	}

	response := struct {
		OriginalNumber int    `json:"original_number"`
		EncodedResult  string `json:"encoded_result"`
		DecodedResult  int    `json:"decoded_result"`
	}{
		OriginalNumber: originalNumber,
		EncodedResult:  resp.Result,
		DecodedResult:  decodedResult,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
