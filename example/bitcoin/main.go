package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"

	"btc_example/lib"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/joho/godotenv"
)

func main() {
	// 환경 설정
	network := "regtest"
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	privateKey := os.Getenv("PRIVATE_KEY")

	// ********************************
	// 키생성 요청
	fmt.Printf("\n############################\n")
	fmt.Printf("\n 1. Key Generation: DKG를 사용한 키 생성 단계 \n\n")
	var keyGenResp *lib.KeyGenResponse
	keyGenFilePath := "key_gen_response.json"

	if _, err := os.Stat(keyGenFilePath); os.IsNotExist(err) {
		keyGenResp, err = performKeyGen()
		if err != nil {
			log.Fatalf("Key generation failed: %v", err)
		}
		saveKeyGenResponse(keyGenResp)
	} else {
		keyGenResp, err = loadKeyGenResponse(keyGenFilePath)
		if err != nil {
			log.Fatalf("Failed to load existing key: %v", err)
		}
		fmt.Println("Loaded existing key from file.")
	}
	fmt.Printf("Address: %s\n", keyGenResp.Address)
	fmt.Printf("\n############################\n")

	// 2. 테스트 BTC 주입
	amount := big.NewInt(100000000) // 1 BTC
	injectTxHash, err := injectTestBTC(privateKey, keyGenResp.Address, amount, network)
	if err != nil {
		log.Fatalf("Failed to inject test BTC: %v", err)
	}
	fmt.Printf("Inject Transaction Hash: %s\n", injectTxHash)

	// 3. 미서명 트랜잭션 생성
	toAddress := "tb1qt2y5mv8zl65h3lpvmpjrqw9l0axskms574zjz5"
	unsignedTx, utxos, err := createUnsignedTransaction(keyGenResp.Address, toAddress, amount, network)
	if err != nil {
		log.Fatalf("Failed to create unsigned transaction: %v", err)
	}

	// 4. 트랜잭션 해시 계산
	txHash, err := calculateTransactionHash(unsignedTx, utxos, network)
	if err != nil {
		log.Fatalf("Failed to calculate transaction hash: %v", err)
	}

	// 5. 서명 수행
	signResp, err := performSign(keyGenResp.Address, txHash)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	// 6. 서명된 트랜잭션 생성
	signedTx, err := combineSignature(unsignedTx, signResp, keyGenResp.PublicKey, txHash)
	if err != nil {
		log.Fatalf("Failed to combine signature with transaction: %v", err)
	}

	// 7. 서명된 트랜잭션 전송
	signedTxHex, err := signedTxToHex(signedTx)
	if err != nil {
		log.Fatalf("Failed to convert signed transaction to hex: %v", err)
	}

	sendSignedTxHash, err := lib.SendSignedTransaction(signedTxHex, network)
	if err != nil {
		log.Fatalf("Failed to send signed transaction: %v", err)
	}

	fmt.Printf("TxHash: %s \n", sendSignedTxHash)
}

func signedTxToHex(tx *wire.MsgTx) (string, error) {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", fmt.Errorf("failed to serialize transaction: %v", err)
	}
	return hex.EncodeToString(buf.Bytes()), nil
}
func injectTestBTC(privateKey, toAddress string, amount *big.Int, network string) (string, error) {
	return lib.InjectTestBTC(privateKey, toAddress, amount, network)
}

func createUnsignedTransaction(fromAddress, toAddress string, amount *big.Int, network string) (*wire.MsgTx, []lib.UTXO, error) {
	tx, utxos, _, err := lib.CreateUnsignedTransaction(fromAddress, toAddress, amount, network)
	return tx, utxos, err
}

func calculateTransactionHash(tx *wire.MsgTx, utxos []lib.UTXO, network string) ([]byte, error) {
	var params *chaincfg.Params
	switch network {
	case "mainnet":
		params = &chaincfg.MainNetParams
	case "testnet":
		params = &chaincfg.TestNet3Params
	case "regtest":
		params = &chaincfg.RegressionNetParams
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
	return lib.CalculateTransactionHash(tx, utxos, params)
}

func combineSignature(tx *wire.MsgTx, signResp *lib.SignResponse, pubKeyHex string, txHash []byte) (*wire.MsgTx, error) {
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %v", err)
	}
	return lib.CombineBTCUnsignedTxWithSignature(tx, *signResp, pubKeyBytes, txHash)
}

func sendSignedTransaction(tx *wire.MsgTx, network string) (string, error) {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", fmt.Errorf("failed to serialize transaction: %v", err)
	}
	txHex := hex.EncodeToString(buf.Bytes())
	return lib.SendSignedTransaction(txHex, network)
}

func performKeyGen() (*lib.KeyGenResponse, error) {
	reqData := lib.KeyGenRequest{Network: 3} // bitcoin regtest
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %v", err)
	}

	resp, err := http.Post("http://localhost:8080/key_gen", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("HTTP POST request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Data lib.KeyGenResponse `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	return &response.Data, nil
}

func performSign(address string, txOrigin []byte) (*lib.SignResponse, error) {
	signReqData := lib.SignRequest{
		Address:  address,
		TxOrigin: base64.StdEncoding.EncodeToString(txOrigin),
	}

	jsonData, err := json.Marshal(signReqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sign request data: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/sign", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP POST request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sign response body: %v", err)
	}

	var response struct {
		Data lib.SignResponse `json:"data"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sign JSON response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error: %s", string(body))
	}
	return &response.Data, nil

}

func saveKeyGenResponse(resp *lib.KeyGenResponse) {

	file, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal response to JSON: %v", err)
		return
	}

	err = ioutil.WriteFile("key_gen_response.json", file, 0644)
	if err != nil {
		log.Printf("Failed to write response to file: %v", err)
		return
	}

	filepath.Abs("key_gen_response.json")
}

func loadKeyGenResponse(filePath string) (*lib.KeyGenResponse, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %v", err)
	}

	var response lib.KeyGenResponse
	err = json.Unmarshal(file, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON from key file: %v", err)
	}

	return &response, nil
}
