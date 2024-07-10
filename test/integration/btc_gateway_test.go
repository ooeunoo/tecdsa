package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	test_util "tecdsa/test/utils"
	"testing"
	"time"

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

func TestBTCKeyGenIntegration(t *testing.T) {
	// 컨텍스트 생성
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 요청 데이터 생성
	requestData := map[string]interface{}{
		"network": 2, // 2 for Bitcoin TestNet, 1 for Bitcoin MainNet
	}
	jsonData, err := json.Marshal(requestData)
	assert.NoError(t, err)

	// 요청 생성
	req, err := http.NewRequestWithContext(ctx, "POST", gatewayURL+"/key_gen", bytes.NewBuffer(jsonData))
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
func TestBTCSignIntegration(t *testing.T) {
	// 키 생성 필수
	if keyGenResponse == nil {
		t.Fatal("KeyGen response is nil. Run TestKeyGenIntegration first.")
	}

	// TX 생성
	_, txOrigin, err := test_util.GenerateBTCTxOrigin("0xcE2Cf674623E1469153948223113B0951C4302D0")
	assert.NoError(t, err)

	// TX 로그
	t.Logf("tx_origin (hex): %s", hex.EncodeToString(txOrigin))
	t.Logf("tx_origin (base64): %s", base64.StdEncoding.EncodeToString(txOrigin))

	// 요청 폼
	signRequest := map[string]interface{}{
		"address":    keyGenResponse["address"],
		"secret_key": keyGenResponse["secret_key"],
		"tx_origin":  base64.StdEncoding.EncodeToString(txOrigin),
	}

	jsonData, err := json.Marshal(signRequest)
	assert.NoError(t, err)

	// 요청 로그
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

	// txWithSignature, _ := test_util.CombineBTCUnsignedTxWithSignature(tx, response, pub)
	// test_util.PrintBTCSignedTxAsJSON(txWithSignature)
	// assert.NoError(t, err, "Failed to combine unsigned transaction with signature")

	// // 서명된 트랜잭션 검증
	// assert.NotNil(t, txWithSignature, "Signed transaction should not be nil")

	// // 트랜잭션 필드 검증
	// assert.Equal(t, tx.To().Hex(), txWithSignature.To().Hex(), "To address should match")
	// assert.Equal(t, tx.Value().String(), txWithSignature.Value().String(), "Transaction value should match")
	// assert.Equal(t, tx.Gas(), txWithSignature.Gas(), "Gas limit should match")
	// assert.Equal(t, tx.GasPrice().String(), txWithSignature.GasPrice().String(), "Gas price should match")
	// assert.Equal(t, tx.Nonce(), txWithSignature.Nonce(), "Nonce should match")

	// // 서명 검증
	// v, r, s := txWithSignature.RawSignatureValues()
	// assert.NotNil(t, v, "V value should not be nil")
	// assert.NotNil(t, r, "R value should not be nil")
	// assert.NotNil(t, s, "S value should not be nil")

	// // 서명 길이 검증
	// assert.Equal(t, 32, len(r.Bytes()), "R should be 32 bytes long")
	// assert.Equal(t, 32, len(s.Bytes()), "S should be 32 bytes long")

	// // 체인 ID 검증 (예: 메인넷의 경우 1)
	// chainID := txWithSignature.ChainId()
	// assert.Equal(t, big.NewInt(1), chainID, "Chain ID should be 1 for mainnet")

	// // 서명자 복구 및 검증
	// signer := types.NewEIP155Signer(chainID)
	// recoveredAddr, err := types.Sender(signer, txWithSignature)
	// assert.NoError(t, err, "Failed to recover signer address")
	// assert.Equal(t, keyGenResponse["address"], recoveredAddr.Hex(), "Recovered address should match the key generation address")

	// test_util.PrintETHSignedTxAsJSON(txWithSignature)

	// t.Logf("Signed transaction successfully verified")
}
