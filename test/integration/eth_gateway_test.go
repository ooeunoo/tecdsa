package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/stretchr/testify/assert"
)

var (
	gatewayURL           string
	httpClient           *http.Client
	client               *ethclient.Client
	signer               types.EIP155Signer
	tx                   *types.Transaction
	keyGenResponse       map[string]interface{}
	signResponse         map[string]interface{}
	signedRawTransaction string
)

func init() {
	gatewayURL = "http://localhost:8080"

	// HTTP 클라이언트 설정
	httpClient = &http.Client{
		Timeout: time.Second * 30,
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
	if keyGenResponse == nil {
		t.Fatal("KeyGen response is nil. Run TestKeyGenIntegration first.")
	}

	data, _ := keyGenResponse["data"].(map[string]interface{})
	fromAddress, _ := data["address"].(string)

	// TX 생성
	client, _ = ethclient.Dial("https://gateway.tenderly.co/public/sepolia")
	chainID, _ := client.NetworkID(context.Background())

	signer = types.NewEIP155Signer(chainID)
	from := common.HexToAddress(fromAddress)
	to := common.HexToAddress("0xFDcBF476B286796706e273F86aC51163DA737FA8")
	amount := big.NewInt(100000000000000000) // 0.01 Ether
	var txOrigin []byte
	tx, txOrigin = GenerateUnSignedTx(*client, signer, from, to, amount)

	// TX 로그
	t.Logf("tx_origin (hex): %s", hex.EncodeToString(txOrigin))
	t.Logf("tx_origin (base64): %s", base64.StdEncoding.EncodeToString(txOrigin))

	// 요청 폼
	signRequest := map[string]interface{}{
		"address":   fromAddress,
		"tx_origin": base64.StdEncoding.EncodeToString(txOrigin),
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
	t.Logf("Status Code: %d", resp.StatusCode)
	t.Logf("Response Body: %s", string(body))

	// 상태 코드 확인
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// JSON 파싱
	err = json.Unmarshal(body, &signResponse)
	assert.NoError(t, err)

	t.Logf("Parsed Sign Response: %+v", signResponse)

}

func TestETHCombineTxWithSignature(t *testing.T) {
	// 서명 필수
	if signResponse == nil {
		t.Fatal("Sign response is nil. Run TestETHSignIntegration first.")
	}

	// data 필드 접근
	signature, ok := signResponse["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Failed to get data from sign response")
	}

	// V, R, S 값을 추출
	v, _ := signature["v"].(float64)
	rStr, _ := signature["r"].(string)
	sStr, _ := signature["s"].(string)

	vInt := int(v)
	rBytes, _ := base64.StdEncoding.DecodeString(rStr)
	sBytes, _ := base64.StdEncoding.DecodeString(sStr)

	signedTx, _ := tx.WithSignature(signer, append(rBytes, append(sBytes, byte(vInt))...))
	rawTxBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		t.Fatalf("Failed to serialize signed transaction: %v", err)
	}

	signedRawTransaction = hex.EncodeToString(rawTxBytes)
	t.Logf("Signed Transaction Raw Hex: %s", signedRawTransaction)
}

func TestETHSendRawTransaction(t *testing.T) {
	// 서명 필수
	if signedRawTransaction == "" {
		t.Fatal("Sign Raw Transaction is nil. Run TestETHCombineTxWithSignatureAndSend first.")
	}

	signedRawTransactionBytes, _ := hex.DecodeString(signedRawTransaction)

	var decodedSignedRawTransaction *types.Transaction
	rlp.DecodeBytes(signedRawTransactionBytes, &decodedSignedRawTransaction)

	client.SendTransaction(context.Background(), decodedSignedRawTransaction)

}

// 트랜잭션 객체 생성
func CreateNewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *types.Transaction {
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    amount,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})
	return tx
}

func GenerateUnSignedTx(client ethclient.Client, signer types.Signer, fromAddress common.Address, toAddress common.Address, amount *big.Int) (*types.Transaction, []byte) {
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("failed to get nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("failed to get gas price: %v", err)
	}
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(15))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(10))

	gasLimit := uint64(21000)
	tx := CreateNewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)
	rlpEncodedTx, _ := rlp.EncodeToBytes([]interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.To(),
		tx.Value(),
		tx.Data(),
		signer.ChainID(), uint(0), uint(0),
	})
	return tx, rlpEncodedTx
}
