package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"
	eth_test_util "tecdsa/test/utils"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

var (
	gatewayURL     string
	keyGenResponse map[string]interface{}
	httpClient     *http.Client
)

func init() {
	gatewayURL = os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080" // 기본값 설정
	}

	// HTTP 클라이언트 설정
	httpClient = &http.Client{
		Timeout: time.Second * 30, // 30초 타임아웃 설정
	}
}

func TestKeyGenIntegration(t *testing.T) {
	// 컨텍스트 생성
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 요청 생성
	req, err := http.NewRequestWithContext(ctx, "POST", gatewayURL+"/key_gen", nil)
	assert.NoError(t, err)

	// 요청 전송
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if resp != nil {
		defer resp.Body.Close()
	} else {
		t.Fatal("Response is nil")
	}

	// 응답 본문 읽기
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	// 응답 로깅
	t.Logf("Status Code: %d", resp.StatusCode)
	t.Logf("Response Body: %s", string(body))

	// 상태 코드 확인
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// JSON 파싱
	err = json.Unmarshal(body, &keyGenResponse)
	assert.NoError(t, err)

	// 응답 필드 확인
	assert.True(t, keyGenResponse["success"].(bool))
	assert.NotEmpty(t, keyGenResponse["address"])
	assert.NotEmpty(t, keyGenResponse["secret_key"])
	assert.Greater(t, keyGenResponse["duration"].(float64), 0.0)

	// 파싱된 응답 로깅
	t.Logf("Parsed KeyGen Response: %+v", keyGenResponse)
}
func TestSignIntegration(t *testing.T) {
	// Ensure we have a KeyGen response
	if keyGenResponse == nil {
		t.Fatal("KeyGen response is nil. Run TestKeyGenIntegration first.")
	}

	// Generate test tx_origin
	tx, txOrigin, err := eth_test_util.GenerateTxOrigin("0xcE2Cf674623E1469153948223113B0951C4302D0")
	assert.NoError(t, err)

	// Log tx_origin details
	t.Logf("tx_origin (hex): %s", hex.EncodeToString(txOrigin))
	t.Logf("tx_origin (base64): %s", base64.StdEncoding.EncodeToString(txOrigin))

	// Prepare the sign request
	signRequest := map[string]interface{}{
		"address":    keyGenResponse["address"],
		"secret_key": keyGenResponse["secret_key"],
		"tx_origin":  base64.StdEncoding.EncodeToString(txOrigin), // Send as base64 encoded string
	}

	jsonData, err := json.Marshal(signRequest)
	assert.NoError(t, err)

	// Log the request body
	t.Logf("Sign Request: %s", string(jsonData))

	// 컨텍스트 생성
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 요청 생성
	req, err := http.NewRequestWithContext(ctx, "POST", gatewayURL+"/sign", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// 요청 전송
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if resp != nil {
		defer resp.Body.Close()
	} else {
		t.Fatal("Response is nil")
	}

	// 응답 본문 읽기
	response, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	// 응답 로깅
	t.Logf("Response Body: %s", string(response))

	// 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}
	type SignResponse struct {
		Success bool   `json:"success"`
		V       string `json:"v"`
		R       string `json:"r"`
		S       string `json:"s"`
	}

	// JSON 파싱
	var signResponse SignResponse
	err = json.Unmarshal(response, &signResponse)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// V를 int64로 변환
	v, err := strconv.ParseInt(signResponse.V, 10, 64)
	if err != nil {
		t.Fatalf("Failed to parse 'v' as int64: %v", err)
	}

	// R과 S를 big.Int로 변환
	rBytes, err := base64.StdEncoding.DecodeString(signResponse.R)
	if err != nil {
		t.Fatalf("Failed to decode 'r': %v", err)
	}
	r := new(big.Int).SetBytes(rBytes)

	sBytes, err := base64.StdEncoding.DecodeString(signResponse.S)
	if err != nil {
		t.Fatalf("Failed to decode 's': %v", err)
	}
	s := new(big.Int).SetBytes(sBytes)

	// 서명 생성 (r + s + v)
	signatureBytes := append(r.Bytes(), s.Bytes()...)
	signatureBytes = append(signatureBytes, byte(v))

	t.Logf("Generated Signature: %x", signatureBytes)
	t.Logf("Signature length: %d", len(signatureBytes))

	signer := types.NewEIP155Signer(big.NewInt(1))
	txWithSignature, _ := tx.WithSignature(signer, signatureBytes)
	eth_test_util.PrintSignedTxAsJSON(txWithSignature)

	// 응답 필드 확인
	success := signResponse.Success
	assert.True(t, success)

	assert.NotEmpty(t, signResponse.V)
	assert.NotEmpty(t, signResponse.R)
	assert.NotEmpty(t, signResponse.S)

	// 파싱된 응답 로깅
	t.Logf("Parsed Sign Response: %+v", signResponse)
}
