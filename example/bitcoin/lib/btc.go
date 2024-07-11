package lib

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func GetUnspentTxs(address string) ([]UTXO, error) {
	url := fmt.Sprintf("https://mempool.space/testnet/api/address/%s/utxo", address)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get UTXOs: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// 디버깅을 위해 응답 내용 출력
	fmt.Printf("API Response: %s\n", string(body))

	var utxos []UTXO
	err = json.Unmarshal(body, &utxos)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal UTXOs: %v", err)
	}

	return utxos, nil
}

func GetPayToAddrScript(address string, network *chaincfg.Params) ([]byte, error) {
	addr, err := btcutil.DecodeAddress(address, network)
	if err != nil {
		return nil, fmt.Errorf("failed to decode address: %v", err)
	}
	return txscript.PayToAddrScript(addr)
}

func InjectTestBTC(privateKey string, toAddress string, amount *big.Int) (string, error) {
	wif, err := btcutil.DecodeWIF(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode WIF: %v", err)
	}

	pubKeyHash := btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed())
	fromAddress, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.TestNet3Params)
	if err != nil {
		return "", fmt.Errorf("failed to get from address: %v", err)
	}

	fmt.Println("From Address: ", fromAddress.EncodeAddress())
	unspentTxs, err := GetUnspentTxs(fromAddress.EncodeAddress())
	if err != nil {
		return "", fmt.Errorf("failed to get unspent transactions: %v", err)
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	var totalInput int64
	for _, utxo := range unspentTxs {
		if !utxo.Status.Confirmed {
			continue
		}
		hash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return "", fmt.Errorf("failed to parse txid: %v", err)
		}
		outPoint := wire.NewOutPoint(hash, utxo.Vout)
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)
		totalInput += utxo.Value
		if totalInput >= amount.Int64()+5000 { // 예상 수수료를 고려하여 충분한 입력을 추가
			break
		}
	}

	if totalInput < amount.Int64() {
		return "", fmt.Errorf("insufficient funds")
	}

	toAddr, err := btcutil.DecodeAddress(toAddress, &chaincfg.TestNet3Params)
	if err != nil {
		return "", fmt.Errorf("failed to decode to address: %v", err)
	}
	pkScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		return "", fmt.Errorf("failed to create pkScript: %v", err)
	}
	tx.AddTxOut(wire.NewTxOut(amount.Int64(), pkScript))

	estimatedSize := tx.SerializeSize() + 100
	feeRate := int64(20) // 수수료율을 20 satoshi/byte로 설정
	fee := int64(estimatedSize) * feeRate

	minFee := int64(2202) // 최소 수수료를 2202 satoshi로 설정
	if fee < minFee {
		fee = minFee
	}

	if totalInput < amount.Int64()+fee {
		return "", fmt.Errorf("insufficient funds: have %d satoshis, need %d satoshis", totalInput, amount.Int64()+fee)
	}

	changeAmount := totalInput - amount.Int64() - fee
	if changeAmount > 546 { // 더스트 한계
		changePkScript, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			return "", fmt.Errorf("failed to create change pkScript: %v", err)
		}
		tx.AddTxOut(wire.NewTxOut(changeAmount, changePkScript))
	} else {
		// 잔액이 더스트 한계보다 작으면 수수료에 추가
		fee += changeAmount
	}

	// 수수료가 입력 금액의 50%를 초과하지 않도록 합니다.
	if fee > totalInput/2 {
		return "", fmt.Errorf("fee is too high: %d satoshis", fee)
	}

	for i, txIn := range tx.TxIn {
		utxo := unspentTxs[i]
		witnessProgram, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			return "", fmt.Errorf("failed to create witness program: %v", err)
		}

		// 여기서 PrevOutputFetcher를 생성합니다.
		prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(witnessProgram, utxo.Value)

		// 수정된 NewTxSigHashes 호출
		sigHashes := txscript.NewTxSigHashes(tx, prevOutputFetcher)

		witness, err := txscript.WitnessSignature(tx, sigHashes, i, utxo.Value, witnessProgram, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return "", fmt.Errorf("failed to create witness signature: %v", err)
		}
		txIn.Witness = witness
	}
	for i, _ := range tx.TxIn {
		utxo := unspentTxs[i]
		witnessProgram, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			return "", fmt.Errorf("failed to create witness program for validation: %v", err)
		}

		prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(witnessProgram, utxo.Value)
		sigHashes := txscript.NewTxSigHashes(tx, prevOutputFetcher)

		vm, err := txscript.NewEngine(
			witnessProgram,
			tx,
			i,
			txscript.StandardVerifyFlags,
			nil, // SigCache can be nil
			sigHashes,
			utxo.Value,
			prevOutputFetcher,
		)
		if err != nil {
			return "", fmt.Errorf("failed to create script engine for input %d: %v", i, err)
		}

		if err := vm.Execute(); err != nil {
			return "", fmt.Errorf("failed to validate transaction for input %d: %v", i, err)
		}
	}

	fmt.Printf("Total input: %d satoshis\n", totalInput)
	fmt.Printf("Output amount: %d satoshis\n", amount.Int64())
	fmt.Printf("Fee: %d satoshis\n", fee)
	fmt.Printf("Change: %d satoshis\n", changeAmount)

	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", fmt.Errorf("failed to serialize transaction: %v", err)
	}

	rawTx := hex.EncodeToString(buf.Bytes())
	fmt.Printf("Raw Transaction: %s\n", rawTx)

	err = SendSignedTransaction(rawTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %v", err)
	}

	return tx.TxHash().String(), nil
}

func SendSignedTransaction(signedTxHex string) error {
	url := "https://mempool.space/testnet/api/tx"
	payload := []byte(signedTxHex)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to broadcast transaction: %s", body)
	}

	fmt.Printf("Transaction broadcast successful. Transaction ID: %s\n", body)
	return nil
}
func GetBalance(address string) (int64, error) {
	utxos, err := GetUnspentTxs(address)
	if err != nil {
		return 0, fmt.Errorf("failed to get UTXOs: %v", err)
	}

	var balance int64
	for _, utxo := range utxos {
		balance += utxo.Value
	}

	return balance, nil
}
