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

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/joho/godotenv"
)

func loadENV() string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv("PRIVATE_KEY")
}

func main() {
	PRIVATE_KEY := loadENV()
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
	// 테스트 BTC 주입
	// fmt.Printf("\n 2. Inject Test BTC: 이후 코인 전송 테스트를 위한 테스트 비트 주입 단계  \n\n")
	// toAddress := keyGenResp.Address
	// amount := big.NewInt(1000) // 0.00001 BTC in satoshis
	// txHash, err := lib.InjectTestBTC(PRIVATE_KEY, toAddress, amount)
	// if err != nil {
	// 	log.Fatalf("Failed to inject test BTC: %v", err)
	// }
	// fmt.Printf("TxHash: %s \n", txHash)
	// fmt.Println("Waiting for block confirmations...")
	// err = lib.WaitForConfirmations(txHash)
	// if err != nil {
	// 	log.Fatalf("Failed to wait for confirmations: %v", err)
	// }
	// fmt.Println("Transaction confirmed")
	// fmt.Printf("\n############################\n")

	// ********************************
	// 서명 데이터 생성
	fmt.Printf("\n 3. Create Encoded Unsigned Transaction: 서명되지않은 트랜잭션 데이터 생성 단계 \n\n")

	amount2 := big.NewInt(1000) // 0.00001 BTC in satoshis
	tx, _, _, err := lib.CreateUnsignedTransaction(keyGenResp.Address, "tb1qt2y5mv8zl65h3lpvmpjrqw9l0axskms574zjz5", amount2)
	if err != nil {
		log.Fatalf("Failed to create unsigned transaction: %v", err)
	}
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		log.Fatalf("Failed to serialize transaction: %v", err)
	}
	encodedBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	fmt.Printf("encoded Transaction(base64): %s\n", encodedBase64)
	fmt.Printf("\n############################\n")

	// ********************************
	// 서명 요청
	fmt.Printf("\n 4. Sign Transaction: 서명 단계 \n\n")
	signResp, err := performSign(keyGenResp.Address, keyGenResp.SecretKey, buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}
	fmt.Println("v: ", signResp.V)
	fmt.Println("r: ", signResp.R)
	fmt.Println("s: ", signResp.S)
	fmt.Printf("\n############################\n")

	// ********************************
	// 트랜잭션, 서명 결합
	signedTx, err := lib.CombineBTCUnsignedTxWithSignature(tx, *signResp)
	if err != nil {
		log.Fatalf("Failed to combine unsigned transaction with signature: %v", err)
	}
	var signedBuf bytes.Buffer
	if err := signedTx.Serialize(&signedBuf); err != nil {
		log.Fatalf("Failed to serialize signed transaction: %v", err)
	}

	signedTxHex := hex.EncodeToString(signedBuf.Bytes())
	fmt.Println("txHash:", signedTx.TxHash())
	if err != nil {
		err = lib.PrintBTCTransactionInfo(signedTxHex)
		if err != nil {
			log.Printf("Failed to print transaction info: %v", err)
		}
	}
	fmt.Printf("\n 5. Signed Transaction Detail: 트랜잭션과 서명 결합 후 완성된 Raw Transaction \n%s\n", signedTxHex)
	fmt.Printf("\n############################\n")

	// ********************************
	// 네트워크 전파
	fmt.Printf("\n############################\n")
	fmt.Printf("\n 6. Send Test BTC Using Sign Signature: 네트워크 전파\n\n")
	txHash, _ := lib.SendSignedTransaction(signedTxHex)
	if err != nil {
		log.Fatalf("Failed to send signed transaction: %v", err)
	}
	fmt.Printf("TxHash: %s \n", txHash)
	fmt.Printf("\n############################\n")
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

// func SignBitcoinTransaction(tx *wire.MsgTx, utxos []lib.UTXO, signFunc func([]byte) ([]byte, error)) error {
// 	for i, txIn := range tx.TxIn {
// 		utxo := utxos[i]
// 		prevOutScript, err := txscript.PayToAddrScript(utxo.Address)
// 		if err != nil {
// 			return fmt.Errorf("failed to create prevOutScript: %v", err)
// 		}

// 		hash, err := CreateSignatureHash(tx, i, prevOutScript, txscript.SigHashAll, utxo.Value)
// 		if err != nil {
// 			return fmt.Errorf("failed to create signature hash: %v", err)
// 		}

// 		signature, err := signFunc(hash)
// 		if err != nil {
// 			return fmt.Errorf("failed to sign input %d: %v", i, err)
// 		}

// 		// Append SIGHASH_ALL byte
// 		signature = append(signature, byte(txscript.SigHashAll))

// 		// Create witness
// 		txIn.Witness = wire.TxWitness{signature, utxo.Address.ScriptAddress()}
// 	}
// 	return nil
// }

// func CreateSignatureHash(tx *wire.MsgTx, idx int, prevOutScript []byte, hashType txscript.SigHashType, amount int64) ([]byte, error) {
// 	return txscript.CalcWitnessSigHash(prevOutScript, txscript.NewTxSigHashes(tx), hashType, tx, idx, amount)
// }
