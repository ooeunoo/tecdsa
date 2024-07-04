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

	// 문자열을 정수로 변환
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

	// Base64 디코딩
	decodedBytes, err := base64.StdEncoding.DecodeString(resp.Result)
	if err != nil {
		http.Error(w, "Error decoding result", http.StatusInternalServerError)
		return
	}

	// 디코딩된 바이트를 정수로 변환
	decodedResult, err := strconv.Atoi(string(decodedBytes))
	if err != nil {
		http.Error(w, "Error converting result to integer", http.StatusInternalServerError)
		return
	}

	// JSON 응답 생성
	response := struct {
		OriginalNumber int `json:"original_number"`
		DecodedResult  int `json:"decoded_result"`
	}{
		OriginalNumber: originalNumber,
		DecodedResult:  decodedResult,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
