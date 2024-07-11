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
	test_util "tecdsa/test/utils"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/stretchr/testify/assert"
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

func TestETHKeyGenIntegration(t *testing.T) {
	// 컨텍스트 생성
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// 요청 데이터 생성
	requestData := map[string]interface{}{
		"network": 4, // 3 for Ethereum Mainnet
	}
	jsonData, err := json.Marshal(requestData)
	assert.NoError(t, err)

	// 요청 생성
	req, err := http.NewRequestWithContext(ctx, "POST", gatewayURL+"/key_gen", bytes.NewBuffer(jsonData))
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

	t.Logf("Parsed KeyGen Response: %+v", keyGenResponse)
}
func TestETHSignIntegration(t *testing.T) {
	// 키 생성 필수
	if keyGenResponse == nil {
		t.Fatal("KeyGen response is nil. Run TestKeyGenIntegration first.")
	}

	// TX 생성
	tx, txOrigin, err := test_util.GenerateETHTxOrigin("0xcE2Cf674623E1469153948223113B0951C4302D0")
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
	t.Logf("Sign Request Data: %+v", signRequest)

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
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	// 응답 로깅
	t.Logf("Response Body: %s", string(body))

	// 상태 코드 확인
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// JSON 파싱
	var signResponse struct {
		Data test_util.SignResponse `json:"data"`
	}
	err = json.Unmarshal(body, &signResponse)
	assert.NoError(t, err)

	// 응답 필드 확인
	assert.True(t, signResponse.Data.Success)
	assert.NotZero(t, signResponse.Data.V)
	assert.NotZero(t, signResponse.Data.R)
	assert.NotZero(t, signResponse.Data.S)

	// V, R, S 값을 big.Int로 변환
	v := new(big.Int).SetUint64(signResponse.Data.V)
	r := new(big.Int).SetBytes(signResponse.Data.R)
	s := new(big.Int).SetBytes(signResponse.Data.S)

	// 서명된 트랜잭션 생성
	chainID := big.NewInt(1) // 메인넷의 경우
	signer := types.NewEIP155Signer(chainID)

	// R과 S를 32바이트로 패딩
	rBytes := make([]byte, 32)
	sBytes := make([]byte, 32)
	r.FillBytes(rBytes)
	s.FillBytes(sBytes)

	signedTx, err := tx.WithSignature(signer, append(rBytes, append(sBytes, v.Bytes()...)...))
	assert.NoError(t, err, "Failed to create signed transaction")

	// 서명된 트랜잭션 검증
	assert.NotNil(t, signedTx, "Signed transaction should not be nil")

	// 트랜잭션 필드 검증
	assert.Equal(t, tx.To().Hex(), signedTx.To().Hex(), "To address should match")
	assert.Equal(t, tx.Value().String(), signedTx.Value().String(), "Transaction value should match")
	assert.Equal(t, tx.Gas(), signedTx.Gas(), "Gas limit should match")
	assert.Equal(t, tx.GasPrice().String(), signedTx.GasPrice().String(), "Gas price should match")
	assert.Equal(t, tx.Nonce(), signedTx.Nonce(), "Nonce should match")

	// 서명 검증
	vv, rr, ss := signedTx.RawSignatureValues()
	assert.Equal(t, v.Uint64(), vv.Uint64(), "V value should match")
	assert.Equal(t, r.Bytes()[0], rr.Bytes()[0], "R value should match")
	assert.Equal(t, s.Bytes()[0], ss.Bytes()[0], "S value should match")

	// 서명자 복구 및 검증
	recoveredAddr, err := types.Sender(signer, signedTx)
	assert.NoError(t, err, "Failed to recover signer address")
	assert.Equal(t, keyGenResponse["address"], recoveredAddr.Hex(), "Recovered address should match the key generation address")

	test_util.PrintETHSignedTxAsJSON(signedTx)

	t.Logf("Signed transaction successfully verified")

}
