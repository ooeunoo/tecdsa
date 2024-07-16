package lib

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func GetUnspentTxs(address string, network string) ([]UTXO, error) {
	var url string
	switch network {
	case "mainnet":
		url = fmt.Sprintf("https://mempool.space/api/address/%s/utxo", address)
	case "testnet":
		url = fmt.Sprintf("https://mempool.space/testnet/api/address/%s/utxo", address)
	case "regtest":
		return getRegtestUnspentTxs(address)
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get UTXOs: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var utxos []UTXO
	err = json.Unmarshal(body, &utxos)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal UTXOs: %v", err)
	}

	return utxos, nil
}

func getRegtestUnspentTxs(address string) ([]UTXO, error) {
	rpcURL := "http://localhost:18443"
	rpcUser := "myuser"
	rpcPassword := "SomeDecentp4ssw0rd"

	payload := []byte(fmt.Sprintf(`{
        "jsonrpc": "1.0",
        "id": "curltest",
        "method": "listunspent",
        "params": [1, 9999999, ["%s"], true, {
            "minimumAmount": 0,
            "maximumCount": 1000
        }]
    }`, address))

	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(rpcUser, rpcPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var rpcResponse struct {
		Result []struct {
			TxID          string  `json:"txid"`
			Vout          uint32  `json:"vout"`
			Address       string  `json:"address"`
			Amount        float64 `json:"amount"`
			Confirmations int64   `json:"confirmations"`
			ScriptPubKey  string  `json:"scriptPubKey"`
			Spendable     bool    `json:"spendable"`
			Solvable      bool    `json:"solvable"`
			Safe          bool    `json:"safe"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &rpcResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	var utxos []UTXO
	for _, unspent := range rpcResponse.Result {
		scriptPubKey, err := hex.DecodeString(unspent.ScriptPubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode scriptPubKey: %v", err)
		}
		utxos = append(utxos, UTXO{
			TxID:         unspent.TxID,
			Vout:         unspent.Vout,
			Address:      unspent.Address,
			Value:        int64(unspent.Amount * 1e8), // BTC to satoshis
			ScriptPubKey: scriptPubKey,
			Status: UTXOStatus{
				Confirmed:   unspent.Confirmations > 0,
				BlockHeight: 0,  // This information is not provided by listunspent
				BlockHash:   "", // This information is not provided by listunspent
				BlockTime:   0,  // This information is not provided by listunspent
			},
		})
	}

	return utxos, nil
}
func InjectTestBTC(privateKey string, toAddress string, amount *big.Int, network string) (string, error) {
	var params *chaincfg.Params
	switch network {
	case "mainnet":
		params = &chaincfg.MainNetParams
	case "testnet":
		params = &chaincfg.TestNet3Params
	case "regtest":
		params = &chaincfg.RegressionNetParams
	default:
		return "", fmt.Errorf("unsupported network: %s", network)
	}

	wif, err := btcutil.DecodeWIF(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode WIF: %v", err)
	}

	pubKey := wif.PrivKey.PubKey()
	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())

	var fromAddress btcutil.Address
	var isP2WPKH bool

	// Try to create a P2WPKH address first
	fromAddress, err = btcutil.NewAddressPubKeyHash(pubKeyHash, params)
	// if err == nil {
	// 	isP2WPKH = true
	// } else {
	// 	// If P2WPKH fails, create a P2PKH address
	// 	fromAddress, err = btcutil.NewAddressPubKeyHash(pubKeyHash, params)
	// 	if err != nil {
	// 		return "", fmt.Errorf("failed to get from address: %v", err)
	// 	}
	// 	isP2WPKH = false
	// }
	// fmt.Println("isP2WPKH:", isP2WPKH)
	// fmt.Println("here address:", fromAddress)

	tx, unspentTxs, _, err := CreateUnsignedTransaction(fromAddress.EncodeAddress(), toAddress, amount, network)
	if err != nil {
		return "", err
	}

	for i, txIn := range tx.TxIn {
		utxo := unspentTxs[i]
		if isP2WPKH {
			witnessProgram, err := txscript.PayToAddrScript(fromAddress)
			if err != nil {
				return "", fmt.Errorf("failed to create witness program: %v", err)
			}

			prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(witnessProgram, utxo.Value)
			sigHashes := txscript.NewTxSigHashes(tx, prevOutputFetcher)

			witness, err := txscript.WitnessSignature(tx, sigHashes, i, utxo.Value, witnessProgram, txscript.SigHashAll, wif.PrivKey, true)
			if err != nil {
				return "", fmt.Errorf("failed to create witness signature: %v", err)
			}

			txIn.Witness = witness
			txIn.SignatureScript = nil
		} else {
			pkScript, err := txscript.PayToAddrScript(fromAddress)
			if err != nil {
				return "", fmt.Errorf("failed to create pkScript: %v", err)
			}

			signature, err := txscript.SignatureScript(tx, i, pkScript, txscript.SigHashAll, wif.PrivKey, true)
			if err != nil {
				return "", fmt.Errorf("failed to create signature script: %v", err)
			}

			txIn.SignatureScript = signature
			txIn.Witness = nil
		}
	}

	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", fmt.Errorf("failed to serialize transaction: %v", err)
	}

	rawTx := hex.EncodeToString(buf.Bytes())
	fmt.Printf("Raw Transaction: %s\n", rawTx)

	err = PrintTransactionInfo(rawTx)
	if err != nil {
		fmt.Printf("Failed to print transaction info: %v\n", err)
	}

	txHash, err := SendSignedTransaction(rawTx, network)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %v", err)
	}

	return txHash, nil
}
func SendSignedTransaction(signedTxHex string, network string) (string, error) {
	var url string
	switch network {
	case "mainnet":
		url = "https://mempool.space/api/tx"
	case "testnet":
		url = "https://mempool.space/testnet/api/tx"
	case "regtest":
		return sendRegtestSignedTransaction(signedTxHex)
	default:
		return "", fmt.Errorf("unsupported network: %s", network)
	}

	payload := []byte(signedTxHex)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to broadcast transaction. Status: %d, Body: %s", resp.StatusCode, string(body))
	}

	txHash := string(body)
	fmt.Printf("Transaction broadcast successful. Transaction ID: %s\n", txHash)

	return txHash, nil
}

func sendRegtestSignedTransaction(signedTxHex string) (string, error) {
	rpcURL := "http://localhost:18443"
	rpcUser := "myuser"
	rpcPassword := "SomeDecentp4ssw0rd"

	payload := []byte(fmt.Sprintf(`{
        "jsonrpc": "1.0",
        "id": "curltest",
        "method": "sendrawtransaction",
        "params": ["%s"]
    }`, signedTxHex))

	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(rpcUser, rpcPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var rpcResponse struct {
		Result string `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	err = json.Unmarshal(body, &rpcResponse)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if rpcResponse.Error != nil {
		return "", fmt.Errorf("RPC error: %d - %s", rpcResponse.Error.Code, rpcResponse.Error.Message)
	}

	return rpcResponse.Result, nil
}

func GetBalance(address string, network string) (int64, error) {
	utxos, err := GetUnspentTxs(address, network)
	if err != nil {
		return 0, fmt.Errorf("failed to get UTXOs: %v", err)
	}

	var balance int64
	for _, utxo := range utxos {
		balance += utxo.Value
	}

	return balance, nil
}

func PrintTransactionInfo(rawTxHex string) error {
	rawTxBytes, err := hex.DecodeString(rawTxHex)
	if err != nil {
		return fmt.Errorf("failed to decode raw transaction: %v", err)
	}

	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(rawTxBytes))
	if err != nil {
		return fmt.Errorf("failed to deserialize transaction: %v", err)
	}

	fmt.Printf("Transaction ID: %s\n", tx.TxHash().String())
	fmt.Printf("Version: %d\n", tx.Version)
	fmt.Printf("Locktime: %d\n", tx.LockTime)

	fmt.Printf("Inputs (%d):\n", len(tx.TxIn))
	for i, txIn := range tx.TxIn {
		fmt.Printf("  Input %d:\n", i)
		fmt.Printf("    Previous Output: %s\n", txIn.PreviousOutPoint.String())
		fmt.Printf("    Sequence: %d\n", txIn.Sequence)
	}

	fmt.Printf("Outputs (%d):\n", len(tx.TxOut))
	for i, txOut := range tx.TxOut {
		fmt.Printf("  Output %d:\n", i)
		fmt.Printf("    Value: %d satoshis\n", txOut.Value)
		fmt.Printf("    Script: %x\n", txOut.PkScript)
	}

	return nil
}
func CreateUnsignedTransaction(fromAddress string, toAddress string, amount *big.Int, network string) (*wire.MsgTx, []UTXO, int64, error) {
	// 네트워크 파라미터 설정
	var params *chaincfg.Params
	switch network {
	case "mainnet":
		params = &chaincfg.MainNetParams
	case "testnet":
		params = &chaincfg.TestNet3Params
	case "regtest":
		params = &chaincfg.RegressionNetParams
	default:
		return nil, nil, 0, fmt.Errorf("unsupported network: %s", network)
	}

	// UTXO 가져오기
	unspentTxs, err := GetUnspentTxs(fromAddress, network)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get unspent transactions: %v", err)
	}

	// 새 트랜잭션 생성
	tx := wire.NewMsgTx(wire.TxVersion)

	// 입력 추가 및 총 입력 금액 계산
	var totalInput int64
	for _, utxo := range unspentTxs {
		if !utxo.Status.Confirmed {
			continue
		}
		hash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to parse txid: %v", err)
		}
		outPoint := wire.NewOutPoint(hash, utxo.Vout)
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)
		totalInput += utxo.Value
		if totalInput >= amount.Int64() {
			break
		}
	}

	// 자금이 충분한지 확인
	if totalInput < amount.Int64() {
		return nil, nil, 0, fmt.Errorf("insufficient funds: have %d satoshis, need %d satoshis", totalInput, amount.Int64())
	}

	// 출력 주소 디코딩
	toAddr, err := btcutil.DecodeAddress(toAddress, params)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to decode to address: %v", err)
	}

	// 출력 스크립트 생성
	pkScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to create pkScript: %v", err)
	}

	// 출력 추가
	tx.AddTxOut(wire.NewTxOut(amount.Int64(), pkScript))

	// 수수료 계산
	estimatedSize := tx.SerializeSize() + 100 // 서명을 위한 추가 공간
	feeRate := int64(20)                      // satoshis per byte
	fee := int64(estimatedSize) * feeRate
	minFee := int64(2202) // 최소 수수료
	if fee < minFee {
		fee = minFee
	}

	// 변경 금액 계산 및 처리
	changeAmount := totalInput - amount.Int64() - fee
	if changeAmount > 546 { // 더스트 한계
		fromAddr, err := btcutil.DecodeAddress(fromAddress, params)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to decode from address: %v", err)
		}
		changePkScript, err := txscript.PayToAddrScript(fromAddr)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to create change pkScript: %v", err)
		}
		tx.AddTxOut(wire.NewTxOut(changeAmount, changePkScript))
	} else {
		fee += changeAmount // 작은 변경 금액은 수수료에 추가
	}

	// 최종 수수료 확인
	if fee > totalInput/2 {
		return nil, nil, 0, fmt.Errorf("fee is too high: %d satoshis", fee)
	}

	return tx, unspentTxs, fee, nil
}

func SignTransaction(tx *wire.MsgTx, unspentTxs []UTXO, wif *btcutil.WIF, fromAddress btcutil.Address) error {
	for i, txIn := range tx.TxIn {
		// utxo := unspentTxs[i]
		pkScript, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			return fmt.Errorf("failed to create pkScript: %v", err)
		}

		signature, err := txscript.SignatureScript(tx, i, pkScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return fmt.Errorf("failed to create signature script: %v", err)
		}

		txIn.SignatureScript = signature
	}
	return nil
}

func ValidateTransaction(tx *wire.MsgTx, unspentTxs []UTXO, fromAddress btcutil.Address) error {
	for i, _ := range tx.TxIn {
		utxo := unspentTxs[i]
		pkScript, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			return fmt.Errorf("failed to create pkScript for validation: %v", err)
		}

		prevOutFetcher := txscript.NewCannedPrevOutputFetcher(pkScript, utxo.Value)
		sigCache := txscript.NewSigCache(10)
		sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)

		vm, err := txscript.NewEngine(
			pkScript,
			tx,
			i,
			txscript.StandardVerifyFlags,
			sigCache,
			sigHashes,
			utxo.Value,
			prevOutFetcher,
		)
		if err != nil {
			return fmt.Errorf("failed to create script engine for input %d: %v", i, err)
		}

		if err := vm.Execute(); err != nil {
			return fmt.Errorf("failed to validate transaction for input %d: %v", i, err)
		}
	}
	return nil
}

func WaitForConfirmations(txHash string, network string) error {
	var url string
	switch network {
	case "mainnet":
		url = fmt.Sprintf("https://mempool.space/api/tx/%s", txHash)
	case "testnet":
		url = fmt.Sprintf("https://mempool.space/testnet/api/tx/%s", txHash)
	case "regtest":
		return nil // Regtest doesn't need to wait for confirmations
	default:
		return fmt.Errorf("unsupported network: %s", network)
	}

	for {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to get transaction info: %v", err)
		}
		defer resp.Body.Close()

		var txInfo struct {
			Status struct {
				Confirmed   bool   `json:"confirmed"`
				BlockHeight int    `json:"block_height"`
				BlockHash   string `json:"block_hash"`
				BlockTime   int    `json:"block_time"`
			} `json:"status"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&txInfo); err != nil {
			return fmt.Errorf("failed to decode transaction info: %v", err)
		}

		if txInfo.Status.Confirmed {
			return nil
		}

		time.Sleep(1 * time.Minute) // Wait for 1 minute before checking again
	}
}
func VerifySignature(tx *wire.MsgTx, idx int, prevOutScript []byte, signature []byte, pubKey []byte) (bool, error) {
	sigHash, err := CalculateSignatureHash(tx, idx, prevOutScript, txscript.SigHashAll)
	if err != nil {
		return false, err
	}

	pubKeyObj, err := btcec.ParsePubKey(pubKey)
	if err != nil {
		return false, err
	}

	sigObj, err := ecdsa.ParseSignature(signature)
	if err != nil {
		return false, err
	}

	return sigObj.Verify(sigHash, pubKeyObj), nil
}
func VerifyTransactionSignature(tx *wire.MsgTx, signResponse SignResponse, pubKeyBytes []byte, hashToSign []byte) (bool, error) {
	rBytes, err := base64.StdEncoding.DecodeString(signResponse.R)
	if err != nil {
		return false, fmt.Errorf("failed to decode R: %v", err)
	}
	sBytes, err := base64.StdEncoding.DecodeString(signResponse.S)
	if err != nil {
		return false, fmt.Errorf("failed to decode S: %v", err)
	}

	r := new(btcec.ModNScalar)
	if overflow := r.SetByteSlice(rBytes); overflow {
		return false, fmt.Errorf("R value is too large")
	}
	s := new(btcec.ModNScalar)
	if overflow := s.SetByteSlice(sBytes); overflow {
		return false, fmt.Errorf("S value is too large")
	}

	signature := ecdsa.NewSignature(r, s)

	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %v", err)
	}

	fmt.Printf("Verifying signature for transaction hash: %x\n", hashToSign)
	fmt.Printf("Using public key: %x\n", pubKeyBytes)

	isValid := signature.Verify(hashToSign, pubKey)
	fmt.Printf("Signature verification result: %v\n", isValid)

	return isValid, nil
}

func CombineBTCUnsignedTxWithSignature(unsignedTx *wire.MsgTx, signResponse SignResponse, pubKeyBytes []byte, hashToSign []byte) (*wire.MsgTx, error) {
	fmt.Println("Combining unsigned transaction with signature...")

	rBytes, err := base64.StdEncoding.DecodeString(signResponse.R)
	if err != nil {
		return nil, fmt.Errorf("failed to decode R: %v", err)
	}
	sBytes, err := base64.StdEncoding.DecodeString(signResponse.S)
	if err != nil {
		return nil, fmt.Errorf("failed to decode S: %v", err)
	}

	signature := append(rBytes, sBytes...)
	signature = append(signature, byte(txscript.SigHashAll))

	fmt.Printf("Full signature: %x\n", signature)
	fmt.Printf("R: %x\n", rBytes)
	fmt.Printf("S: %x\n", sBytes)
	fmt.Printf("Public key: %x\n", pubKeyBytes)

	isValid, err := VerifyTransactionSignature(unsignedTx, signResponse, pubKeyBytes, hashToSign)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %v", err)
	}
	if !isValid {
		return nil, fmt.Errorf("invalid signature")
	}

	signedTx := wire.NewMsgTx(unsignedTx.Version)

	for _, txOut := range unsignedTx.TxOut {
		signedTx.AddTxOut(txOut)
	}

	for _, txIn := range unsignedTx.TxIn {
		signedTxIn := wire.NewTxIn(&txIn.PreviousOutPoint, nil, nil)

		builder := txscript.NewScriptBuilder()
		builder.AddData(signature)
		builder.AddData(pubKeyBytes)
		sigScript, err := builder.Script()
		if err != nil {
			return nil, fmt.Errorf("failed to create signature script: %v", err)
		}
		signedTxIn.SignatureScript = sigScript

		signedTx.AddTxIn(signedTxIn)
	}

	fmt.Println("Transaction signing completed")
	return signedTx, nil
}

func CalculateSignatureHash(tx *wire.MsgTx, idx int, prevOutScript []byte, hashType txscript.SigHashType) ([]byte, error) {
	return txscript.CalcSignatureHash(prevOutScript, hashType, tx, idx)
}

func CreateScriptSig(signature []byte, pubKey []byte) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddData(signature)
	builder.AddData(pubKey)
	return builder.Script()
}

func CalculateTransactionHash(tx *wire.MsgTx, utxos []UTXO, params *chaincfg.Params) ([]byte, error) {
	// 트랜잭션의 복사본을 만듭니다.
	txCopy := tx.Copy()

	// 모든 입력의 SignatureScript를 비웁니다.
	for _, txIn := range txCopy.TxIn {
		txIn.SignatureScript = nil
	}

	// 각 입력에 대해 서명 해시를 계산합니다.
	var sigHashes [][]byte
	for i, _ := range txCopy.TxIn {
		utxo := utxos[i]

		// UTXO의 주소를 디코딩합니다.
		addr, err := btcutil.DecodeAddress(utxo.Address, params)
		if err != nil {
			return nil, fmt.Errorf("failed to decode address: %v", err)
		}

		// 주소에 대한 스크립트를 생성합니다.
		prevOutScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to create output script: %v", err)
		}

		// 서명 해시를 계산합니다.
		hash, err := txscript.CalcSignatureHash(prevOutScript, txscript.SigHashAll, txCopy, i)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate signature hash: %v", err)
		}

		sigHashes = append(sigHashes, hash)
	}

	// 모든 서명 해시를 결합합니다.
	var combinedHash []byte
	for _, hash := range sigHashes {
		combinedHash = append(combinedHash, hash...)
	}

	// 결합된 해시의 SHA256을 계산합니다.
	finalHash := chainhash.DoubleHashB(combinedHash)

	return finalHash, nil
}
