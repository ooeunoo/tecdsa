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

	tx, unspentTxs, _, err := CreateUnsignedTransaction(fromAddress.EncodeAddress(), toAddress, amount)
	if err != nil {
		return "", err
	}

	// 서명 과정
	err = SignTransaction(tx, unspentTxs, wif, fromAddress)
	if err != nil {
		return "", err
	}

	// 트랜잭션 유효성 검사
	err = ValidateTransaction(tx, unspentTxs, fromAddress)
	if err != nil {
		return "", err
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

func CreateUnsignedTransaction(fromAddress string, toAddress string, amount *big.Int) (*wire.MsgTx, []UTXO, int64, error) {
	unspentTxs, err := GetUnspentTxs(fromAddress)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get unspent transactions: %v", err)
	}

	tx := wire.NewMsgTx(wire.TxVersion)
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
	}

	if totalInput < amount.Int64() {
		return nil, nil, 0, fmt.Errorf("insufficient funds: have %d satoshis, need %d satoshis", totalInput, amount.Int64())
	}

	toAddr, err := btcutil.DecodeAddress(toAddress, &chaincfg.TestNet3Params)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to decode to address: %v", err)
	}
	pkScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to create pkScript: %v", err)
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
		return nil, nil, 0, fmt.Errorf("insufficient funds: have %d satoshis, need %d satoshis", totalInput, amount.Int64()+fee)
	}

	changeAmount := totalInput - amount.Int64() - fee
	if changeAmount > 546 { // 더스트 한계
		fromAddr, err := btcutil.DecodeAddress(fromAddress, &chaincfg.TestNet3Params)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to decode from address: %v", err)
		}
		changePkScript, err := txscript.PayToAddrScript(fromAddr)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to create change pkScript: %v", err)
		}
		tx.AddTxOut(wire.NewTxOut(changeAmount, changePkScript))
	} else {
		// 잔액이 더스트 한계보다 작으면 수수료에 추가
		fee += changeAmount
	}

	// 수수료가 입력 금액의 50%를 초과하지 않도록 합니다.
	if fee > totalInput/2 {
		return nil, nil, 0, fmt.Errorf("fee is too high: %d satoshis", fee)
	}

	return tx, unspentTxs, fee, nil
}

func SignTransaction(tx *wire.MsgTx, unspentTxs []UTXO, wif *btcutil.WIF, fromAddress btcutil.Address) error {
	for i, txIn := range tx.TxIn {
		utxo := unspentTxs[i]
		witnessProgram, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			return fmt.Errorf("failed to create witness program: %v", err)
		}

		prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(witnessProgram, utxo.Value)
		sigHashes := txscript.NewTxSigHashes(tx, prevOutputFetcher)

		witness, err := txscript.WitnessSignature(tx, sigHashes, i, utxo.Value, witnessProgram, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return fmt.Errorf("failed to create witness signature: %v", err)
		}
		txIn.Witness = witness
	}
	return nil
}

func ValidateTransaction(tx *wire.MsgTx, unspentTxs []UTXO, fromAddress btcutil.Address) error {
	for i, _ := range tx.TxIn {
		utxo := unspentTxs[i]
		witnessProgram, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			return fmt.Errorf("failed to create witness program for validation: %v", err)
		}

		prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(witnessProgram, utxo.Value)
		sigHashes := txscript.NewTxSigHashes(tx, prevOutputFetcher)

		vm, err := txscript.NewEngine(
			witnessProgram,
			tx,
			i,
			txscript.StandardVerifyFlags,
			nil,
			sigHashes,
			utxo.Value,
			prevOutputFetcher,
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

func WaitForConfirmations(txHash string) error {
	for {
		url := fmt.Sprintf("https://mempool.space/testnet/api/tx/%s", txHash)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to get transaction info: %v", err)
		}
		defer resp.Body.Close()

		type TxStatus struct {
			Confirmed   bool   `json:"confirmed"`
			BlockHeight int    `json:"block_height"`
			BlockHash   string `json:"block_hash"`
			BlockTime   int    `json:"block_time"`
		}
		var TxInfo struct {
			Status TxStatus `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&TxInfo); err != nil {
			return fmt.Errorf("failed to decode transaction info: %v", err)
		}

		if TxInfo.Status.Confirmed {
			return nil
		}

		time.Sleep(1 * time.Minute) // Wait for 1 minute before checking again
	}
}
func CombineBTCUnsignedTxWithSignature(unsignedTx *wire.MsgTx, signResponse SignResponse) (*wire.MsgTx, error) {
	// Decode R and S from base64
	rBytes, err := base64.StdEncoding.DecodeString(signResponse.R)
	if err != nil {
		return nil, fmt.Errorf("failed to decode R: %v", err)
	}
	sBytes, err := base64.StdEncoding.DecodeString(signResponse.S)
	if err != nil {
		return nil, fmt.Errorf("failed to decode S: %v", err)
	}

	// Create signature bytes
	signatureBytes := append(rBytes, sBytes...)

	// Add sighash type
	signatureBytes = append(signatureBytes, byte(txscript.SigHashAll))

	// Create a new transaction to hold the signed version
	signedTx := wire.NewMsgTx(unsignedTx.Version)

	// Copy TxOut from unsigned transaction
	for _, txOut := range unsignedTx.TxOut {
		signedTx.AddTxOut(txOut)
	}

	// Sign each input
	for _, txIn := range unsignedTx.TxIn {
		signedTxIn := wire.NewTxIn(&txIn.PreviousOutPoint, nil, nil)

		// For P2WPKH, we need to create a witness
		// But we don't have the public key at this point, so we'll use a placeholder
		witness := wire.TxWitness{signatureBytes, []byte{}} // Empty byte slice as placeholder for pubkey
		signedTxIn.Witness = witness

		signedTx.AddTxIn(signedTxIn)
	}

	return signedTx, nil
}

func PrintBTCTransactionInfo(rawTxHex string) error {
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
		fmt.Printf(" Input %d:\n", i)
		fmt.Printf("  Previous Output: %s\n", txIn.PreviousOutPoint.String())
		fmt.Printf("  Sequence: %d\n", txIn.Sequence)
	}

	fmt.Printf("Outputs (%d):\n", len(tx.TxOut))
	for i, txOut := range tx.TxOut {
		fmt.Printf(" Output %d:\n", i)
		fmt.Printf("  Value: %d satoshis\n", txOut.Value)
		fmt.Printf("  Script: %x\n", txOut.PkScript)
	}

	return nil
}
