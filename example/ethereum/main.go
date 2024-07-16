package main

import (
	"bytes"
	"context"
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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func loadENV() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv("INFURA_KEY"), os.Getenv("PRIVATE_KEY")
}

func main() {
	INFURA_KEY, PRIVATE_KEY := loadENV()

	infuraURL := fmt.Sprintf("https://sepolia.infura.io/v3/%s", INFURA_KEY)
	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Fatalf("failed to connect to the Ethereum client: %v", err)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

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
	fmt.Printf("ParitalSecretShare Key: %s\n", keyGenResp.SecretKey)
	fmt.Printf("\n############################\n")

	// ********************************
	// 테스트 ETH 주입
	fmt.Printf("\n 2. Inject Test Ether: 이후 코인 전송 테스트를 위한 테스트 이더 주입 단계  \n\n")
	amount := big.NewInt(100000000000000000) // 0.01 Ether
	receipt, _ := lib.InjectTestEther(client, PRIVATE_KEY, keyGenResp.Address, amount)
	fmt.Printf("TxHash: %s \n", receipt.TxHash)
	fmt.Printf("Transaction confirmed in block %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("\n############################\n")

	// ********************************
	// 서명 데이터 생성
	fmt.Printf("\n 3. Create Encoded Rlp Transaction: 트랜잭션 RLP 인코딩 데이터 생성 단계 \n\n")
	signer := types.NewEIP155Signer(chainID)
	tx, rlpEncodedTx := lib.GenerateRlpEncodedTx(
		*client,
		signer,
		common.HexToAddress(keyGenResp.Address),
		common.HexToAddress("0xFDcBF476B286796706e273F86aC51163DA737FA8"),
		new(big.Int).Div(amount, big.NewInt(3)),
	)
	encodedBase64 := base64.StdEncoding.EncodeToString(rlpEncodedTx)
	fmt.Printf("encoded Transaction(base64): %s\n", encodedBase64)
	fmt.Printf("\n############################\n")

	// ********************************
	// 서명 요청
	fmt.Printf("\n 4. Sign Transaction: 서명 단계 \n\n")
	signResp, _ := performSign(keyGenResp.Address, keyGenResp.SecretKey, rlpEncodedTx)
	fmt.Println("v: ", signResp.V)
	fmt.Println("r: ", signResp.R)
	fmt.Println("s: ", signResp.S)
	fmt.Printf("\n############################\n")

	// ********************************
	// 트랜잭션, 서명 결합
	signedTx, err := lib.CombineETHUnsignedTxWithSignature(tx, chainID, *signResp)
	if err != nil {
		log.Fatalf("Failed to combine unsigned transaction with signature: %v", err)
	}
	data, _ := lib.PrintETHSignedTxAsJSON(signedTx)
	fmt.Printf("\n 5. Signed Transaction Detail: 트랜잭션과 서명 결합 후 완성된 Raw Transaction \n%s\n", string(data))

	// ********************************
	// 네트워크 전파
	fmt.Printf("\n############################\n")
	fmt.Printf("\n 6. Send Test Ether Using Sign Signature: 네트워크 전파\n\n")
	receipt2, err := lib.SendSignedTransaction(client, signedTx, true)
	if err != nil {
		log.Fatalf("Failed to send signed transaction: %v", err)
	}
	fmt.Printf("TxHash: %s \n", receipt2.TxHash)
	fmt.Printf("Transaction confirmed in block %d\n", receipt2.BlockNumber.Uint64())
	fmt.Printf("\n############################\n")

}

func performKeyGen() (*lib.KeyGenResponse, error) {
	reqData := lib.KeyGenRequest{Network: 4} // ethereum sepolia
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
