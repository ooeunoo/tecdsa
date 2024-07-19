package lib

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	// 최소 수수료율 (satoshi/byte)
	minFeeRate = 4
	// 예상 트랜잭션 크기 (바이트)
	estimatedTxSize = 250
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

func GetAddressType(address string, params *chaincfg.Params) AddressType {
	addr, err := btcutil.DecodeAddress(address, params)
	if err != nil {
		return Unknown
	}

	switch addr.(type) {
	case *btcutil.AddressPubKeyHash:
		return P2PKH
	case *btcutil.AddressScriptHash:
		return P2SH
	case *btcutil.AddressWitnessPubKeyHash:
		return P2WPKH
	case *btcutil.AddressWitnessScriptHash:
		return P2WSH
	default:
		// Additional check for Bech32 addresses
		if strings.HasPrefix(address, "bc1") || strings.HasPrefix(address, "tb1") {
			if len(address) == 42 {
				return P2WPKH
			} else if len(address) == 62 {
				return P2WSH
			}
		}
		return Unknown
	}
}

func CreateUnsignedTransaction(from, to string, amount int64, network string, publicKey []byte) (*wire.MsgTx, error) {
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
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
	// 주소 디코딩
	fromAddr, err := btcutil.DecodeAddress(from, params)
	if err != nil {
		return nil, fmt.Errorf("failed to decode from address: %v", err)
	}
	toAddr, err := btcutil.DecodeAddress(to, params)
	if err != nil {
		return nil, fmt.Errorf("failed to decode to address: %v", err)
	}

	// UTXO 가져오기
	utxos, err := GetUnspentTxs(from, network)
	if err != nil {
		return nil, fmt.Errorf("failed to get UTXOs: %v", err)
	}

	// 트랜잭션 생성
	tx := wire.NewMsgTx(wire.TxVersion)

	// 수수료 계산 (예상 크기 * 최소 수수료율)
	fee := int64(estimatedTxSize * minFeeRate)

	// 입력 추가
	var totalIn int64
	for _, utxo := range utxos {
		if totalIn >= amount+fee {
			break
		}
		hash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse txid: %v", err)
		}
		outpoint := wire.NewOutPoint(hash, utxo.Vout)
		txIn := wire.NewTxIn(outpoint, nil, nil)
		// txIn.SignatureScript = publicKey // 여기에 공개키 추가
		tx.AddTxIn(txIn)
		totalIn += utxo.Value
	}

	if totalIn < amount+fee {
		return nil, fmt.Errorf("insufficient funds for transaction and fee")
	}

	// 출력 추가 (수신자)
	pkScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pkscript: %v", err)
	}
	tx.AddTxOut(wire.NewTxOut(amount, pkScript))

	// 잔액 처리
	change := totalIn - amount - fee
	if change > 0 {
		changePkScript, err := txscript.PayToAddrScript(fromAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create change pkscript: %v", err)
		}
		tx.AddTxOut(wire.NewTxOut(change, changePkScript))
	}

	return tx, nil
}

func SerializeTransaction(tx *wire.MsgTx) ([]byte, error) {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %v", err)
	}
	return buf.Bytes(), nil
}

func ApplySignature(tx *wire.MsgTx, signResp *SignResponse, pubKeyBytes []byte) error {
	// Decode r and s from base64
	r, err := base64.StdEncoding.DecodeString(signResp.R)
	if err != nil {
		return fmt.Errorf("failed to decode R: %v", err)
	}
	s, err := base64.StdEncoding.DecodeString(signResp.S)
	if err != nil {
		return fmt.Errorf("failed to decode S: %v", err)
	}

	// Create signature from r and s
	rScalar := new(btcec.ModNScalar)
	sScalar := new(btcec.ModNScalar)

	rScalar.SetByteSlice(r)
	sScalar.SetByteSlice(s)

	signature := ecdsa.NewSignature(rScalar, sScalar)

	// Decode public key
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %v", err)
	}

	// Create DER-encoded signature
	sigDER := signature.Serialize()
	sig := append(sigDER, byte(txscript.SigHashAll))

	// Apply signature to each input
	for i := range tx.TxIn {
		sigScript, err := txscript.NewScriptBuilder().
			AddData(sig).
			AddData(pubKey.SerializeCompressed()).
			Script()
		if err != nil {
			return fmt.Errorf("failed to build signature script: %v", err)
		}
		tx.TxIn[i].SignatureScript = sigScript
	}

	return nil
}

func BroadcastTransaction(tx *wire.MsgTx, network string) (string, error) {
	var url string
	switch network {
	case "mainnet":
		url = "https://api.blockcypher.com/v1/btc/main/txs/push"
	case "testnet":
		url = "https://api.blockcypher.com/v1/btc/test3/txs/push"
	case "regtest":
		return broadcastRegtestTransaction(tx)
	default:
		return "", fmt.Errorf("unsupported network: %s", network)
	}

	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", fmt.Errorf("failed to serialize transaction: %v", err)
	}

	payload := struct {
		Tx string `json:"tx"`
	}{
		Tx: hex.EncodeToString(buf.Bytes()),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var result struct {
		Tx struct {
			Hash string `json:"hash"`
		} `json:"tx"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %v", err)
	}

	if result.Tx.Hash == "" {
		return "", fmt.Errorf("transaction hash not found in response: %s", string(body))
	}

	return result.Tx.Hash, nil
}
func broadcastRegtestTransaction(tx *wire.MsgTx) (string, error) {
	rpcURL := "http://localhost:18443"
	rpcUser := "myuser"
	rpcPassword := "SomeDecentp4ssw0rd"

	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", fmt.Errorf("failed to serialize transaction: %v", err)
	}

	hexString := hex.EncodeToString(buf.Bytes())

	payload := []byte(fmt.Sprintf(`{
        "jsonrpc": "1.0",
        "id": "curltest",
        "method": "sendrawtransaction",
        "params": ["%s", 0.1]
    }`, hexString))

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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var result struct {
		Result string      `json:"result"`
		Error  interface{} `json:"error"`
		ID     string      `json:"id"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %v", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("RPC error: %v", result.Error)
	}

	fmt.Printf("Transaction broadcast successful. Transaction hash: %s\n", result.Result)
	return result.Result, nil
}
