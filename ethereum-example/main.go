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

	"ethereum_example/lib"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {}

func loadENV() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv("INFURA_KEY"), os.Getenv("PRIVATE_KEY")
}

func connectToEthereum(infuraKey string) (*ethclient.Client, error) {
	infuraURL := fmt.Sprintf("https://sepolia.infura.io/v3/%s", infuraKey)
	return ethclient.Dial(infuraURL)
}

func handleKeyGeneration() (*lib.KeyGenResponse, error) {
	keyGenFilePath := "key_gen_response.json"
	if _, err := os.Stat(keyGenFilePath); os.IsNotExist(err) {
		return performKeyGen()
	}
	return loadKeyGenResponse(keyGenFilePath)
}

func saveKeyGenResponse(resp *lib.KeyGenResponse) {
	file, _ := json.MarshalIndent(resp, "", " ")
	_ = ioutil.WriteFile("key_gen_response.json", file, 0644)
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

	saveKeyGenResponse(&response.Data)
	return &response.Data, nil
}

func injectTestEther(client *ethclient.Client, privateKey, address string) error {
	amount := big.NewInt(100000000000000000) // 0.1 Ether
	_, err := lib.InjectTestEther(client, privateKey, address, amount)
	return err
}

func signTransaction(address string, rlpEncodedTx []byte) (*lib.SignResponse, error) {
	signReqData := lib.SignRequest{
		Address:  address,
		TxOrigin: base64.StdEncoding.EncodeToString(rlpEncodedTx),
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
