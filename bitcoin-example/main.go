package main

import (
	"bytes"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"sort"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

var network = &chaincfg.RegressionNetParams

const (
	fromAddress        string = "mv8DvRq31mtQkpug6bSjfWitCicLDUuU2i"
	fromAddressPubKey  string = "030a43c5bd559d18c9b3a8dacd16c4ce42dee241d5f858b928460ed3c7594abea3"
	toAddress          string = "mygGWsTEhWRQg8nbwsJ9aPXeDpxdroaTag"
	amount             int64  = 1000000 // 0.01 BTC in satoshis
	fee                int64  = 1000    // 1000 satoshis fee
	remoteSignEndpoint        = "http://localhost:8080/sign"
	rpcURL                    = "http://localhost:18443"
	rpcUser                   = "myuser"
	rpcPassword               = "SomeDecentp4ssw0rd"
)

type SignResponse struct {
	V int    `json:"v"`
	R string `json:"r"`
	S string `json:"s"`
}

type UTXO struct {
	TxID         string `json:"txid"`
	Vout         uint32 `json:"vout"`
	Amount       int64  `json:"amount"`
	ScriptPubKey string `json:"scriptPubKey"` // 문자열로 변경
}

type SignatureResponse struct {
	R string `json:"r"`
	S string `json:"s"`
}

func main() {
	// 1. UTXO 가져오기
	utxos, err := fetchUnspentTxs(fromAddress)
	if err != nil {
		log.Fatalf("Error fetching UTXOs: %v", err)
	}

	// 필요한 총 금액 계산 (송금액 + 수수료)
	requiredAmount := amount + fee

	// 필요한 UTXO 선택
	selectedUTXOs, totalInput, err := selectUTXOs(utxos, requiredAmount)
	if err != nil {
		log.Fatalf("Error selecting UTXOs: %v", err)
	}

	// 선택된 UTXO 출력
	selectedUTXOsJSON, err := json.MarshalIndent(selectedUTXOs, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling selected UTXOs to JSON: %v", err)
	}

	fmt.Printf("Selected UTXOs:\n%s\n", string(selectedUTXOsJSON))
	fmt.Printf("Total input: %d satoshis\n", totalInput)
	fmt.Printf("Amount to send: %d satoshis\n", amount)
	fmt.Printf("Fee: %d satoshis\n", fee)
	fmt.Printf("Change: %d satoshis\n", totalInput-amount-fee)

	redeemTx, signHashes, _ := createTx(fromAddress, toAddress, amount, selectedUTXOs)

	// 각 입력에 대해 개별적으로 서명
	for i, signHash := range signHashes {
		signResponse, err := performSign(fromAddress, signHash)
		if err != nil {
			log.Fatalf("Error signing input %d: %v", i, err)
		}

		// base64 디코딩
		r, err := base64.StdEncoding.DecodeString(signResponse.R)
		if err != nil {
			log.Fatalf("Error decoding R for input %d: %v", i, err)
		}
		s, err := base64.StdEncoding.DecodeString(signResponse.S)
		if err != nil {
			log.Fatalf("Error decoding S for input %d: %v", i, err)
		}

		if err := validateSignatureComponents(r, s, signHash, fromAddressPubKey); err != nil {
			log.Fatalf("Invalid signature for input %d: %v", i, err)
		}

		// DER 인코딩 서명 생성
		signature, err := createDERSignature(r, s)
		if err != nil {
			log.Fatalf("Error creating DER signature for input %d: %v", i, err)
		}

		pubKeyBytes, _ := hex.DecodeString(fromAddressPubKey)
		sigScript, _ := txscript.NewScriptBuilder().
			AddData(signature). // This should already include the hash type
			AddData(pubKeyBytes).
			Script()

		log.Printf("Signing hash: %x", signHash)
		log.Printf("Signature R: %x", r)
		log.Printf("Signature S: %x", s)
		log.Printf("DER Signature: %x", signature)
		log.Printf("sigScript: %x", sigScript)
		log.Printf("UTXO being spent: %+v", selectedUTXOs[i])
		log.Printf("Public Key: %x", pubKeyBytes)
		log.Printf("Full SignatureScript: %x", sigScript)
		redeemTx.TxIn[i].SignatureScript = sigScript
	}

	// 서명된 트랜잭션 직렬화
	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)
	finalRawTx := hex.EncodeToString(signedTx.Bytes())

	fmt.Printf("Final signed transaction: %s\n", finalRawTx)

	err = validateSignedTransaction(finalRawTx, selectedUTXOs, network)
	if err != nil {
		log.Fatalf("Failed validation ", err)
	}

	// 서명된 트랜잭션 브로드캐스트
	tx, _ := broadcastTransaction(redeemTx)
	fmt.Printf("Final broadcasted transaction: %s\n", tx)
}

func fetchUnspentTxs(address string) ([]UTXO, error) {
	payload := []byte(fmt.Sprintf(`{
        "jsonrpc": "1.0",
        "id": "curltest",
        "method": "listunspent",
        "params": [1, 9999999, ["%s"], true, {"minimumAmount": 0, "maximumCount": 1000}]
    }`, address))

	resp, err := makeRPCRequest(payload)
	if err != nil {
		return nil, err
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

	err = json.Unmarshal(resp, &rpcResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	var utxos []UTXO
	for _, unspent := range rpcResponse.Result {
		utxos = append(utxos, UTXO{
			TxID:         unspent.TxID,
			Vout:         unspent.Vout,
			Amount:       int64(unspent.Amount * 1e8), // BTC to satoshis
			ScriptPubKey: unspent.ScriptPubKey,        // 이미 16진수 문자열
		})
	}

	fmt.Println("utxos:", utxos)

	return utxos, nil
}

func selectUTXOs(utxos []UTXO, targetAmount int64) ([]UTXO, int64, error) {
	// UTXO를 금액 기준 내림차순으로 정렬
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Amount > utxos[j].Amount
	})

	var selectedUTXOs []UTXO
	var totalAmount int64

	for _, utxo := range utxos {
		err := verifyUTXO(utxo.TxID, utxo.Vout)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid utxo")
		}
		selectedUTXOs = append(selectedUTXOs, utxo)
		totalAmount += utxo.Amount

		if totalAmount >= targetAmount {
			break
		}
	}

	if totalAmount < targetAmount {
		return nil, 0, fmt.Errorf("insufficient funds: available %d, required %d", totalAmount, targetAmount)
	}

	return selectedUTXOs, totalAmount, nil
}

func verifyUTXO(txid string, vout uint32) error {
	payload := []byte(fmt.Sprintf(`{
        "jsonrpc": "1.0",
        "id": "curltest",
        "method": "gettxout",
        "params": ["%s", %d, true]
    }`, txid, vout))

	resp, err := makeRPCRequest(payload)
	if err != nil {
		return err
	}

	var result struct {
		Result struct {
			ScriptPubKey struct {
				Hex string `json:"hex"`
			} `json:"scriptPubKey"`
			Value float64 `json:"value"`
		} `json:"result"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	log.Printf("Verified UTXO: txid=%s, vout=%d, scriptPubKey=%s, value=%f",
		txid, vout, result.Result.ScriptPubKey.Hex, result.Result.Value)

	return nil
}

func createTx(fromAddress string, toAddress string, amount int64, selectedUTXOs []UTXO) (*wire.MsgTx, [][]byte, error) {

	redeemTx := wire.NewMsgTx(wire.TxVersion)

	var totalInput int64
	for _, utxo := range selectedUTXOs {
		utxoHash, _ := chainhash.NewHashFromStr(utxo.TxID)
		outPoint := wire.NewOutPoint(utxoHash, utxo.Vout)
		txIn := wire.NewTxIn(outPoint, nil, nil)
		redeemTx.AddTxIn(txIn)
		totalInput += utxo.Amount
	}

	// 목적지 출력 추가
	fromAddr, _ := btcutil.DecodeAddress(fromAddress, network)
	fromAddrByte, _ := txscript.PayToAddrScript(fromAddr)

	// 목적지 출력 추가
	destinationAddr, _ := btcutil.DecodeAddress(toAddress, network)
	destinationAddrByte, _ := txscript.PayToAddrScript(destinationAddr)
	redeemTxOut := wire.NewTxOut(amount, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	// 거스름돈 출력 추가 (필요한 경우)
	changeAmount := totalInput - amount - fee
	if changeAmount > 0 {
		changeTxOut := wire.NewTxOut(changeAmount, fromAddrByte)
		redeemTx.AddTxOut(changeTxOut)
	}

	log.Printf("Transaction structure:")
	log.Printf("Version: %d", redeemTx.Version)
	log.Printf("Locktime: %d", redeemTx.LockTime)

	log.Printf("Inputs:")
	for i, txIn := range redeemTx.TxIn {
		log.Printf("  Input %d:", i)
		log.Printf("    PreviousOutPoint: %s", txIn.PreviousOutPoint)
		log.Printf("    Sequence: %d", txIn.Sequence)
	}

	log.Printf("Outputs:")
	for i, txOut := range redeemTx.TxOut {
		log.Printf("  Output %d:", i)
		log.Printf("    Value: %d", txOut.Value)
		log.Printf("    PkScript: %x", txOut.PkScript)
	}

	// 서명을 위한 해시 준비
	signHashes := make([][]byte, len(redeemTx.TxIn))
	for i, _ := range redeemTx.TxIn {
		utxo := selectedUTXOs[i]
		pkScript, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return nil, nil, fmt.Errorf("error decoding pkScript: %v", err)
		}
		signHash, err := txscript.CalcSignatureHash(pkScript, txscript.SigHashAll, redeemTx, i)
		if err != nil {
			return nil, nil, fmt.Errorf("error calculating signature hash: %v", err)
		}
		signHashes[i] = signHash
	}

	return redeemTx, signHashes, nil
}

func makeRPCRequest(payload []byte) ([]byte, error) {
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}

func performSign(address string, sigHash []byte) (*SignResponse, error) {
	fmt.Println("signHash", sigHash)
	signReqData := struct {
		Address  string `json:"address"`
		TxOrigin string `json:"tx_origin"`
	}{
		Address:  address,
		TxOrigin: base64.StdEncoding.EncodeToString(sigHash),
	}

	jsonData, _ := json.Marshal(signReqData)
	req, _ := http.NewRequest("POST", "http://localhost:8080/sign", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP POST request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var response struct {
		Data SignResponse `json:"data"`
	}
	json.Unmarshal(body, &response)

	return &response.Data, nil
}
func createDERSignature(r, s []byte) ([]byte, error) {
	rInt := new(big.Int).SetBytes(r)
	sInt := new(big.Int).SetBytes(s)

	// Ensure low S
	halfOrder := new(big.Int).Rsh(btcec.S256().N, 1)
	if sInt.Cmp(halfOrder) > 0 {
		sInt.Sub(btcec.S256().N, sInt)
	}

	// Create the signature struct
	signature := struct {
		R, S *big.Int
	}{
		R: rInt,
		S: sInt,
	}

	// Marshal to DER encoding
	derSignature, err := asn1.Marshal(signature)
	if err != nil {
		return nil, err
	}

	// Append the hash type (SIGHASH_ALL)
	return append(derSignature, byte(txscript.SigHashAll)), nil
}

func broadcastTransaction(tx *wire.MsgTx) (string, error) {
	var buf bytes.Buffer
	tx.Serialize(&buf)
	txHex := hex.EncodeToString(buf.Bytes())

	rpcURL := "http://localhost:18443"
	rpcUser := "myuser"
	rpcPassword := "SomeDecentp4ssw0rd"

	payload := fmt.Sprintf(`{"jsonrpc":"1.0","id":"curltest","method":"sendrawtransaction","params":["%s"]}`, txHex)
	req, _ := http.NewRequest("POST", rpcURL, bytes.NewBufferString(payload))
	req.SetBasicAuth(rpcUser, rpcPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result struct {
		Result string      `json:"result"`
		Error  interface{} `json:"error"`
	}
	json.Unmarshal(body, &result)

	if result.Error != nil {
		return "", fmt.Errorf("RPC error: %v", result.Error)
	}

	return result.Result, nil
}
func validateSignedTransaction(txHex string, utxos []UTXO, network *chaincfg.Params) error {

	// 16진수 문자열을 바이트 슬라이스로 디코딩
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return fmt.Errorf("error decoding transaction hex: %v", err)
	}

	// 바이트 슬라이스를 wire.MsgTx로 디시리얼라이즈
	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return fmt.Errorf("error deserializing transaction: %v", err)
	}

	// 트랜잭션 구조 검증
	if len(tx.TxIn) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}
	if len(tx.TxOut) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}
	// UTXO 정보를 기반으로 PrevOutputFetcher 생성
	prevOutputFetcher := txscript.NewMultiPrevOutFetcher(nil)
	for _, utxo := range utxos {
		txHash, _ := chainhash.NewHashFromStr(utxo.TxID)
		outPoint := wire.NewOutPoint(txHash, utxo.Vout)

		pkScript, _ := hex.DecodeString(utxo.ScriptPubKey)
		txOut := wire.NewTxOut(utxo.Amount, pkScript)

		prevOutputFetcher.AddPrevOut(*outPoint, txOut)
	}

	// TxSigHashes 생성
	sigHashes := txscript.NewTxSigHashes(&tx, prevOutputFetcher)

	// 각 입력에 대한 서명 검증
	for i, txIn := range tx.TxIn {
		if len(txIn.SignatureScript) == 0 {
			return fmt.Errorf("input %d has no signature script", i)
		}

		// 해당 UTXO 찾기
		var utxo *UTXO
		for _, u := range utxos {
			if u.TxID == txIn.PreviousOutPoint.Hash.String() && u.Vout == txIn.PreviousOutPoint.Index {
				utxo = &u
				break
			}
		}
		if utxo == nil {
			return fmt.Errorf("UTXO not found for input %d", i)
		}

		// 이전 출력 스크립트 디코딩
		prevOutScript, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return fmt.Errorf("error decoding previous output script for input %d: %v", i, err)
		}

		// 서명 검증
		// Create a PrevOutputFetcher
		prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(prevOutScript, utxo.Amount)

		// Create the script engine
		vm, err := txscript.NewEngine(prevOutScript, &tx, i, txscript.StandardVerifyFlags, nil, sigHashes, utxo.Amount, prevOutputFetcher)
		if err != nil {
			return fmt.Errorf("error creating script engine for input %d: %v", i, err)
		}

		log.Printf("Validating input %d", i)
		log.Printf("  PreviousOutPoint: %s", txIn.PreviousOutPoint)
		log.Printf("  SignatureScript: %x", txIn.SignatureScript)
		log.Printf("  UTXO Amount: %d", utxo.Amount)
		log.Printf("  UTXO ScriptPubKey: %s", utxo.ScriptPubKey)

		if err := vm.Execute(); err != nil {
			log.Printf("Script verification failed for input %d: %v", i, err)
			// 스크립트 실행 실패 시 추가 정보 로깅
			log.Printf("  Signature script: %x", txIn.SignatureScript)
			log.Printf("  Public key script: %x", prevOutScript)
			return fmt.Errorf("script verification failed for input %d: %v", i, err)
		}

		log.Printf("Input %d validated successfully", i)
	}

	// 출력 주소 검증
	for i, txOut := range tx.TxOut {
		_, addresses, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, network)
		if err != nil {
			return fmt.Errorf("error extracting address from output %d: %v", i, err)
		}
		if len(addresses) == 0 {
			return fmt.Errorf("no valid address found for output %d", i)
		}
	}

	return nil
}

func validateSignatureComponents(r, s, signHash []byte, pubKeyHex string) error {
	// R과 S를 big.Int로 변환
	rInt := new(big.Int).SetBytes(r)
	sInt := new(big.Int).SetBytes(s)

	// R과 S가 1보다 크고 곡선의 차수보다 작은지 확인
	if rInt.Cmp(big.NewInt(1)) <= 0 || rInt.Cmp(btcec.S256().N) >= 0 {
		return fmt.Errorf("R is out of range")
	}
	if sInt.Cmp(big.NewInt(1)) <= 0 || sInt.Cmp(btcec.S256().N) >= 0 {
		return fmt.Errorf("S is out of range")
	}

	// big.Int를 secp256k1.ModNScalar로 변환
	var rScalar, sScalar btcec.ModNScalar
	rScalar.SetByteSlice(rInt.Bytes())
	sScalar.SetByteSlice(sInt.Bytes())

	// 공개키 파싱
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %v", err)
	}
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %v", err)
	}

	// 서명 재구성
	signature := ecdsa.NewSignature(&rScalar, &sScalar)

	// 서명 검증
	if !signature.Verify(signHash, pubKey) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}
