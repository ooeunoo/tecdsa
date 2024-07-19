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

	"btc_example/lib"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	infuraKey, privateKey := loadENV()
	client, err := connectToEthereum(infuraKey)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum: %v", err)
	}

	keyGenResp, err := handleKeyGeneration()
	if err != nil {
		log.Fatalf("Key generation failed: %v", err)
	}

	err = injectTestEther(client, privateKey, keyGenResp.Address)
	if err != nil {
		log.Fatalf("Failed to inject test ether: %v", err)
	}

	tx, rlpEncodedTx := createEncodedRlpTransaction(client, keyGenResp.Address)
	if err != nil {
		log.Fatalf("Failed to create encoded RLP transaction: %v", err)
	}

	signResp, err := signTransaction(keyGenResp.Address, keyGenResp.SecretKey, rlpEncodedTx)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	signedTx, err := combineAndSendTransaction(client, tx, signResp)
	if err != nil {
		log.Fatalf("Failed to combine and send transaction: %v", err)
	}

	fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
}

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

func injectTestEther(client *ethclient.Client, privateKey, address string) error {
	amount := big.NewInt(100000000000000000) // 0.1 Ether
	_, err := lib.InjectTestEther(client, privateKey, address, amount)
	return err
}

func createEncodedRlpTransaction(client *ethclient.Client, fromAddress string) (*types.Transaction, []byte) {
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, nil
	}

	signer := types.NewEIP155Signer(chainID)
	amount := big.NewInt(33333333333333333) // 0.033333333333333333 Ether
	return lib.GenerateRlpEncodedTx(
		*client,
		signer,
		common.HexToAddress(fromAddress),
		common.HexToAddress("0xFDcBF476B286796706e273F86aC51163DA737FA8"),
		amount,
	)
}

func signTransaction(address, secretKey string, rlpEncodedTx []byte) (*lib.SignResponse, error) {
	signReqData := lib.SignRequest{
		Address:   address,
		SecretKey: secretKey,
		TxOrigin:  base64.StdEncoding.EncodeToString(rlpEncodedTx),
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

func combineAndSendTransaction(client *ethclient.Client, tx *types.Transaction, signResp *lib.SignResponse) (*types.Transaction, error) {
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	signedTx, err := lib.CombineETHUnsignedTxWithSignature(tx, chainID, *signResp)
	if err != nil {
		return nil, err
	}

	_, err = lib.SendSignedTransaction(client, signedTx, true)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}
