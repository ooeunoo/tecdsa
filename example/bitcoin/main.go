package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"

	"btc_example/lib"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/joho/godotenv"
)

func loadENV() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv("TATUM_API_KEY"), os.Getenv("PRIVATE_KEY")
}

func main() {
	_, PRIVATE_KEY := loadENV()
	wif, err := btcutil.DecodeWIF(PRIVATE_KEY)
	if err != nil {
		log.Fatalf("Failed to decode WIF: %v", err)
	}

	pubKeyHash := btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed())
	fromAddress, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.TestNet3Params)
	if err != nil {
		log.Fatalf("Failed to get from address: %v", err)
	}

	balance, err := lib.GetBalance(fromAddress.EncodeAddress())
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	fmt.Printf("Current balance: %d satoshis (%.8f BTC)\n", balance, float64(balance)/100000000)

	// ********************************
	// 키생성 요청
	fmt.Printf("\n############################\n")
	fmt.Printf("\n 1. Key Generation: DKG를 사용한 키 생성 단계 \n\n")
	keyGenResp, err := performKeyGen()
	if err != nil {
		log.Fatalf("Key generation failed: %v", err)
	}
	saveKeyGenResponse(keyGenResp)
	fmt.Printf("Address: %s\n", keyGenResp.Address)
	fmt.Printf("Secret Key: %s\n", keyGenResp.SecretKey)
	fmt.Printf("\n############################\n")

	// ********************************
	// 테스트 BTC 주입
	fmt.Printf("\n 2. Inject Test BTC: 이후 코인 전송 테스트를 위한 테스트 비트 주입 단계  \n\n")
	toAddress := keyGenResp.Address
	amount := big.NewInt(1000) // 0.00001 BTC in satoshis
	txHash, err := lib.InjectTestBTC(PRIVATE_KEY, toAddress, amount)
	if err != nil {
		log.Fatalf("Failed to inject test BTC: %v", err)
	}
	fmt.Printf("TxHash: %s \n", txHash)
	fmt.Printf("\n############################\n")

	// ********************************
	// 서명 데이터 생성
	fmt.Printf("\n 3. Create Encoded Unsigned Transaction: 서명되지않은 트랜잭션 데이터 생성 단계 \n\n")
	// amount2 := big.NewInt(1000) // 0.00001 BTC in satoshis
	// tx, unspentTxs, fee, err := lib.CreateUnsignedTransaction(keyGenResp.Address, "tb1qt2y5mv8zl65h3lpvmpjrqw9l0axskms574zjz5", amount2)

	// fromAddress := keyGenResp.Address
	// toAddress = "tb1qt2y5mv8zl65h3lpvmpjrqw9l0axskms574zjz5"
	// amount = big.NewInt(500) // 0.000005 BTC in satoshis

	// unspentTxs, err := lib.GetUnspentTxs(fromAddress, amount.Int64())
	// if err != nil {
	// 	log.Fatalf("Failed to get unspent transactions: %v", err)
	// }

	// unsignedTx, err := lib.GenerateTransaction(fromAddress, toAddress, amount.Int64(), unspentTxs)
	// if err != nil {
	// 	log.Fatalf("Failed to generate unsigned transaction: %v", err)
	// }

	// // 서명되지 않은 트랜잭션을 개인키로 서명하고 r, s 값 반환
	// r, s, err := lib.SignTransactionByPk(unsignedTx, keyGenResp.SecretKey)
	// if err != nil {
	// 	log.Fatalf("Failed to sign transaction: %v", err)
	// }

	// // 서명되지 않은 트랜잭션과 r, s 값을 결합하여 서명된 트랜잭션 생성
	// signedTx, err := lib.CombineUnsignedTransactionWithSignature(unsignedTx, r, s)
	// if err != nil {
	// 	log.Fatalf("Failed to combine unsigned transaction with signature: %v", err)
	// }

	// // 서명된 트랜잭션을 브로드캐스트
	// err = lib.SendSignedTransaction(signedTx)
	// if err != nil {
	// 	log.Fatalf("Failed to send signed transaction: %v", err)
	// }

	// fmt.Println("Transaction successfully sent!")
}

func performKeyGen() (*lib.KeyGenResponse, error) {
	reqData := lib.KeyGenRequest{Network: 2} // ethereum sepolia
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

func performSign(address string, secretKey string, txOrigin []byte) (*lib.SignResponse, error) {
	signReqData := lib.SignRequest{
		Address:   address,
		SecretKey: secretKey,
		TxOrigin:  base64.StdEncoding.EncodeToString(txOrigin),
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
