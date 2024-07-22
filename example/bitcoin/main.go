package main

import (
	"sync"
)

type KeyGenResponse struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
}

type SignResponse struct {
	V int    `json:"v"`
	R string `json:"r"`
	S string `json:"s"`
}

type UTXO struct {
	TxID         string `json:"txid"`
	Vout         uint32 `json:"vout"`
	Amount       int64  `json:"amount"`
	ScriptPubKey []byte `json:"scriptPubKey"`
}

var (
	utxoCache      map[string][]UTXO
	utxoCacheMutex sync.RWMutex
)

func init() {
	utxoCache = make(map[string][]UTXO)
}

func main() {
	// if err := godotenv.Load(); err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }

	// network := "regtest"
	// keyGenResp, err := loadOrGenerateKey()
	// if err != nil {
	// 	log.Fatalf("Failed to load or generate key: %v", err)
	// }

	// fmt.Printf("Address: %s\n", keyGenResp.Address)

	// // 트랜잭션 생성
	// from := keyGenResp.Address
	// to := "mygGWsTEhWRQg8nbwsJ9aPXeDpxdroaTag"
	// amount := int64(100000000) // 1 BTC in satoshis
	// fee := int64(4000)         // 4000 satoshis

	// unsignedTx, err := createUnsignedTransaction(from, to, amount, fee, network)
	// if err != nil {
	// 	log.Fatalf("Failed to create unsigned transaction: %v", err)
	// }

	// // 트랜잭션 서명
	// signedTx, err := signTransaction(unsignedTx, keyGenResp)
	// if err != nil {
	// 	log.Fatalf("Failed to sign transaction: %v", err)
	// }

	// // 트랜잭션 검증
	// if err := validateTransaction(signedTx, keyGenResp.Address, network); err != nil {
	// 	log.Fatalf("Transaction validation failed: %v", err)
	// }

	// // 트랜잭션 전파
	// txHash, err := broadcastTransaction(signedTx, network)
	// if err != nil {
	// 	log.Fatalf("Failed to broadcast transaction: %v", err)
	// }

	// fmt.Printf("Transaction broadcasted. Hash: %s\n", txHash)
}

// func loadOrGenerateKey() (*KeyGenResponse, error) {
// 	keyGenFilePath := "key_gen_response.json"
// 	if _, err := os.Stat(keyGenFilePath); os.IsNotExist(err) {
// 		return performKeyGen()
// 	}
// 	return loadKeyGenResponse(keyGenFilePath)
// }

// func performKeyGen() (*KeyGenResponse, error) {
// 	reqData := struct {
// 		Network int `json:"network"`
// 	}{
// 		Network: 3, // bitcoin regtest
// 	}
// 	jsonData, err := json.Marshal(reqData)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal request data: %v", err)
// 	}

// 	resp, err := http.Post("http://localhost:8080/key_gen", "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return nil, fmt.Errorf("HTTP POST request failed: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read response body: %v", err)
// 	}

// 	var response struct {
// 		Data KeyGenResponse `json:"data"`
// 	}
// 	err = json.Unmarshal(body, &response)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
// 	}

// 	saveKeyGenResponse(&response.Data)
// 	return &response.Data, nil
// }

// func saveKeyGenResponse(resp *KeyGenResponse) {
// 	file, _ := json.MarshalIndent(resp, "", " ")
// 	_ = ioutil.WriteFile("key_gen_response.json", file, 0644)
// }

// func loadKeyGenResponse(filePath string) (*KeyGenResponse, error) {
// 	file, err := ioutil.ReadFile(filePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read key file: %v", err)
// 	}

// 	var response KeyGenResponse
// 	err = json.Unmarshal(file, &response)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse JSON from key file: %v", err)
// 	}

// 	return &response, nil
// }

// func createUnsignedTransaction(from, to string, amount int64, fee int64, network string) (*wire.MsgTx, error) {
// 	utxos, err := getUnspentTxs(from, network)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get UTXOs: %v", err)
// 	}

// 	tx := wire.NewMsgTx(wire.TxVersion)

// 	var totalIn int64
// 	for _, utxo := range utxos {
// 		if totalIn >= amount+fee {
// 			break
// 		}
// 		hash, _ := chainhash.NewHashFromStr(utxo.TxID)
// 		outPoint := wire.NewOutPoint(hash, utxo.Vout)
// 		txIn := wire.NewTxIn(outPoint, nil, nil)
// 		tx.AddTxIn(txIn)
// 		totalIn += utxo.Amount
// 	}

// 	if totalIn < amount+fee {
// 		return nil, fmt.Errorf("insufficient funds")
// 	}

// 	toAddr, _ := btcutil.DecodeAddress(to, &chaincfg.RegressionNetParams)
// 	pkScript, _ := txscript.PayToAddrScript(toAddr)
// 	tx.AddTxOut(wire.NewTxOut(amount, pkScript))

// 	if totalIn > amount+fee {
// 		changeAddr, _ := btcutil.DecodeAddress(from, &chaincfg.RegressionNetParams)
// 		changePkScript, _ := txscript.PayToAddrScript(changeAddr)
// 		tx.AddTxOut(wire.NewTxOut(totalIn-amount-fee, changePkScript))
// 	}

// 	return tx, nil
// }

// func signTransaction(tx *wire.MsgTx, keyGenResp *KeyGenResponse) (*wire.MsgTx, error) {
// 	pubKeyBytes, err := hex.DecodeString(keyGenResp.PublicKey)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to decode public key: %v", err)
// 	}

// 	for i, txIn := range tx.TxIn {
// 		utxo, err := getUTXO(keyGenResp.Address, txIn.PreviousOutPoint.Hash.String(), txIn.PreviousOutPoint.Index)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to get UTXO: %v", err)
// 		}

// 		sigHash, err := txscript.CalcSignatureHash(utxo.ScriptPubKey, txscript.SigHashAll, tx, i)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to calculate sig hash: %v", err)
// 		}

// 		signResp, err := performSign(keyGenResp.Address, sigHash)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to sign transaction: %v", err)
// 		}

// 		r, _ := base64.StdEncoding.DecodeString(signResp.R)
// 		s, _ := base64.StdEncoding.DecodeString(signResp.S)

// 		// Convert r and s to big.Int
// 		rInt := new(big.Int).SetBytes(r)
// 		sInt := new(big.Int).SetBytes(s)

// 		derSignature, err := encodeSignatureToDER(rInt, sInt)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to encode signature to DER: %v", err)
// 		}

// 		// Add sighash type
// 		signature := append(derSignature, byte(txscript.SigHashAll))

// 		log.Printf("DER-encoded Signature data: %x", signature)
// 		log.Printf("Public key: %x", pubKeyBytes)

// 		builder := txscript.NewScriptBuilder()
// 		builder.AddData(signature)
// 		builder.AddData(pubKeyBytes)
// 		scriptSig, err := builder.Script()
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to build script: %v", err)
// 		}

// 		txIn.SignatureScript = scriptSig
// 	}

// 	return tx, nil
// }

// func encodeSignatureToDER(r, s *big.Int) ([]byte, error) {
// 	var b cryptobyte.Builder
// 	b.AddASN1(asn1.SEQUENCE, func(b *cryptobyte.Builder) {
// 		b.AddASN1BigInt(r)
// 		b.AddASN1BigInt(s)
// 	})
// 	return b.Bytes()
// }

// func performSign(address string, sigHash []byte) (*SignResponse, error) {
// 	signReqData := struct {
// 		Address  string `json:"address"`
// 		TxOrigin string `json:"tx_origin"`
// 	}{
// 		Address:  address,
// 		TxOrigin: base64.StdEncoding.EncodeToString(sigHash),
// 	}

// 	jsonData, _ := json.Marshal(signReqData)
// 	req, _ := http.NewRequest("POST", "http://localhost:8080/sign", bytes.NewBuffer(jsonData))
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("HTTP POST request failed: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, _ := ioutil.ReadAll(resp.Body)

// 	var response struct {
// 		Data SignResponse `json:"data"`
// 	}
// 	json.Unmarshal(body, &response)

// 	return &response.Data, nil
// }

// func broadcastTransaction(tx *wire.MsgTx, network string) (string, error) {
// 	var buf bytes.Buffer
// 	tx.Serialize(&buf)
// 	txHex := hex.EncodeToString(buf.Bytes())

// 	rpcURL := "http://localhost:18443"
// 	rpcUser := "myuser"
// 	rpcPassword := "SomeDecentp4ssw0rd"

// 	payload := fmt.Sprintf(`{"jsonrpc":"1.0","id":"curltest","method":"sendrawtransaction","params":["%s"]}`, txHex)
// 	req, _ := http.NewRequest("POST", rpcURL, bytes.NewBufferString(payload))
// 	req.SetBasicAuth(rpcUser, rpcPassword)
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to send request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, _ := ioutil.ReadAll(resp.Body)
// 	var result struct {
// 		Result string      `json:"result"`
// 		Error  interface{} `json:"error"`
// 	}
// 	json.Unmarshal(body, &result)

// 	if result.Error != nil {
// 		return "", fmt.Errorf("RPC error: %v", result.Error)
// 	}

// 	return result.Result, nil
// }

// func getUnspentTxs(address string, network string) ([]UTXO, error) {
// 	utxos, err := fetchUnspentTxs(address, network)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return utxos, nil
// }

// func fetchUnspentTxs(address string, network string) ([]UTXO, error) {
// 	var url string
// 	switch network {
// 	case "mainnet":
// 		url = fmt.Sprintf("https://mempool.space/api/address/%s/utxo", address)
// 	case "testnet":
// 		url = fmt.Sprintf("https://mempool.space/testnet/api/address/%s/utxo", address)
// 	case "regtest":
// 		return getRegtestUnspentTxs(address)
// 	default:
// 		return nil, fmt.Errorf("unsupported network: %s", network)
// 	}

// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get UTXOs: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read response body: %v", err)
// 	}

// 	var utxos []UTXO
// 	err = json.Unmarshal(body, &utxos)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal UTXOs: %v", err)
// 	}

// 	return utxos, nil
// }

// func getRegtestUnspentTxs(address string) ([]UTXO, error) {
// 	rpcURL := "http://localhost:18443"
// 	rpcUser := "myuser"
// 	rpcPassword := "SomeDecentp4ssw0rd"

// 	payload := []byte(fmt.Sprintf(`{
//         "jsonrpc": "1.0",
//         "id": "curltest",
//         "method": "listunspent",
//         "params": [1, 9999999, ["%s"], true, {"minimumAmount": 0, "maximumCount": 1000}]
//     }`, address))
// 	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(payload))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create request: %v", err)
// 	}

// 	req.SetBasicAuth(rpcUser, rpcPassword)
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to send request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read response body: %v", err)
// 	}

// 	var rpcResponse struct {
// 		Result []struct {
// 			TxID          string  `json:"txid"`
// 			Vout          uint32  `json:"vout"`
// 			Address       string  `json:"address"`
// 			Amount        float64 `json:"amount"`
// 			Confirmations int64   `json:"confirmations"`
// 			ScriptPubKey  string  `json:"scriptPubKey"`
// 			Spendable     bool    `json:"spendable"`
// 			Solvable      bool    `json:"solvable"`
// 			Safe          bool    `json:"safe"`
// 		} `json:"result"`
// 	}

// 	err = json.Unmarshal(body, &rpcResponse)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
// 	}

// 	var utxos []UTXO
// 	for _, unspent := range rpcResponse.Result {
// 		scriptPubKey, err := hex.DecodeString(unspent.ScriptPubKey)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to decode scriptPubKey: %v", err)
// 		}
// 		utxos = append(utxos, UTXO{
// 			TxID:         unspent.TxID,
// 			Vout:         unspent.Vout,
// 			Amount:       int64(unspent.Amount * 1e8), // BTC to satoshis
// 			ScriptPubKey: scriptPubKey,
// 		})
// 	}

// 	return utxos, nil
// }

// func validateTransaction(tx *wire.MsgTx, address string, network string) error {
// 	for i, txIn := range tx.TxIn {
// 		utxo, err := getUTXO(address, txIn.PreviousOutPoint.Hash.String(), txIn.PreviousOutPoint.Index)
// 		if err != nil {
// 			return fmt.Errorf("failed to get UTXO for input %d: %v", i, err)
// 		}

// 		prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(utxo.ScriptPubKey, utxo.Amount)
// 		sigHashes := txscript.NewTxSigHashes(tx, prevOutputFetcher)
// 		vm, err := txscript.NewEngine(utxo.ScriptPubKey, tx, i, txscript.StandardVerifyFlags, nil, sigHashes, utxo.Amount, prevOutputFetcher)
// 		if err != nil {
// 			return fmt.Errorf("failed to create script engine for input %d: %v", i, err)
// 		}

// 		if err := vm.Execute(); err != nil {
// 			return fmt.Errorf("script execution failed for input %d: %v", i, err)
// 		}
// 	}
// 	return nil
// }
