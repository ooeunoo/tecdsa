package network

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"tecdsa/pkg/transaction"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

type BitcoinTxRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"` // satoshi
	Fee    *int64 `json:"fee,omitempty"`
}

type BitcoinOutput struct {
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

type UTXOStatus struct {
	Confirmed   bool   `json:"confirmed"`
	BlockHeight int64  `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	BlockTime   int64  `json:"block_time"`
}

type UTXO struct {
	TxID   string     `json:"txid"`
	Vout   uint32     `json:"vout"`
	Status UTXOStatus `json:"status"`
	Value  int64      `json:"value"`
}

const (
	/*
		P2PKH (Pay-to-Public-Key-Hash):
		설명: 가장 전통적인 비트코인 주소 유형입니다.
		특징:
			주소가 '1'로 시작합니다.
			공개키의 해시를 사용합니다.
			레거시 지갑과 호환성이 좋습니다.
	*/
	P2PKH = 0
	/*
		P2SHP2WPKH (Pay-to-Script-Hash wrapping a Pay-to-Witness-Public-Key-Hash):
		설명: 이는 SegWit 주소를 P2SH 형식으로 감싼 것입니다. 흔히 "nested SegWit" 주소라고 불립니다.
		특징:
			주소가 '3'으로 시작합니다.
			SegWit의 이점을 제공하면서도 이전 지갑과의 호환성을 유지합니다.
	*/
	P2SHP2WPKH = 1
	/*
	   P2WPKH (Pay-to-Witness-Public-Key-Hash):
	   설명: 이는 네이티브 SegWit 주소 유형입니다.
	   특징:
	   		주소가 'bc1'로 시작합니다 (메인넷의 경우).
	   		가장 효율적인 트랜잭션 구조를 제공합니다.
	   		더 낮은 트랜잭션 수수료를 가능하게 합니다.
	*/
	P2WPKH = 2
)

func DeriveBitcoinAddress(point curves.Point, network Network) (string, error) {
	pubKeyBytes := point.ToAffineCompressed()
	if len(pubKeyBytes) == 0 {
		return "", fmt.Errorf("failed to convert public key to bytes")
	}

	var params *chaincfg.Params
	var addrType int
	switch network {
	case Bitcoin:
		params = &chaincfg.MainNetParams
		addrType = P2PKH // 메인넷
	case BitcoinTestNet:
		params = &chaincfg.TestNet3Params
		addrType = P2PKH // 테스트넷
	case BitcoinRegTest:
		params = &chaincfg.RegressionNetParams
		addrType = P2PKH // 로컬
	default:
		return "", fmt.Errorf("unsupported Bitcoin network: %v", network)
	}
	fmt.Println("params:", params.Name)

	var address btcutil.Address
	var err error

	switch addrType {
	case P2PKH:
		hash160 := btcutil.Hash160(pubKeyBytes)
		address, err = btcutil.NewAddressPubKeyHash(hash160, params)
	case P2SHP2WPKH:
		witnessProg := btcutil.Hash160(pubKeyBytes)
		witnessAddress, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, params)
		if err != nil {
			return "", fmt.Errorf("failed to create witness address for P2SH-P2WPKH: %w", err)
		}
		address, err = btcutil.NewAddressScriptHash(witnessAddress.ScriptAddress(), params)
	case P2WPKH:
		witnessProg := btcutil.Hash160(pubKeyBytes)
		address, err = btcutil.NewAddressWitnessPubKeyHash(witnessProg, params)
	default:
		return "", fmt.Errorf("unsupported address type: %d", addrType)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create address: %w", err)
	}

	return address.EncodeAddress(), nil
}

func CreateUnsignedBitcoinTransaction(req interface{}, network Network) (*transaction.UnsignedTransaction, error) {
	btcReq, ok := req.(BitcoinTxRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for Bitcoin transaction")
	}

	var params *chaincfg.Params
	switch network {
	case Bitcoin:
		params = &chaincfg.MainNetParams
	case BitcoinTestNet:
		params = &chaincfg.TestNet3Params
	default:
		return nil, fmt.Errorf("unsupported Bitcoin network: %v", network)
	}

	// Validate addresses
	if !IsValidBitcoinAddress(btcReq.From, network) {
		return nil, fmt.Errorf("invalid 'from' address: %s", btcReq.From)
	}
	if !IsValidBitcoinAddress(btcReq.To, network) {
		return nil, fmt.Errorf("invalid 'to' address: %s", btcReq.To)
	}

	// Amount를 int64로 변환
	amount, ok := new(big.Int).SetString(btcReq.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %s", btcReq.Amount)
	}

	// Get UTXOs for the from address
	utxos, err := GetUnspentTxs(btcReq.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get UTXOs: %v", err)
	}

	fmt.Printf("Found %d UTXOs\n", len(utxos))

	tx := wire.NewMsgTx(wire.TxVersion)

	var totalInput int64
	for _, utxo := range utxos {
		if totalInput >= amount.Int64() {
			break
		}
		hash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse txid: %v", err)
		}
		outPoint := wire.NewOutPoint(hash, utxo.Vout)
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)
		totalInput += utxo.Value
	}

	if totalInput < amount.Int64() {
		return nil, fmt.Errorf("insufficient funds: total input %d is less than required amount %d", totalInput, amount.Int64())
	}

	// Add the output
	toAddress, err := btcutil.DecodeAddress(btcReq.To, params)
	if err != nil {
		return nil, fmt.Errorf("failed to decode to address: %v", err)
	}
	pkScript, err := txscript.PayToAddrScript(toAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create pkScript: %v", err)
	}
	tx.AddTxOut(wire.NewTxOut(amount.Int64(), pkScript))

	// Calculate fee
	estimatedSize := tx.SerializeSize() + 100 // Add some buffer for signatures
	feeRate := int64(20)                      // 20 satoshis per byte
	fee := int64(estimatedSize) * feeRate
	if btcReq.Fee != nil {
		fee = *btcReq.Fee
	}

	if totalInput < amount.Int64()+fee {
		return nil, fmt.Errorf("insufficient funds: have %d satoshis, need %d satoshis", totalInput, amount.Int64()+fee)
	}

	// Add change output if necessary
	changeAmount := totalInput - amount.Int64() - fee
	if changeAmount > 546 { // Dust limit
		fromAddress, err := btcutil.DecodeAddress(btcReq.From, params)
		if err != nil {
			return nil, fmt.Errorf("failed to decode from address: %v", err)
		}
		changePkScript, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create change pkScript: %v", err)
		}
		tx.AddTxOut(wire.NewTxOut(changeAmount, changePkScript))
	} else {
		// 잔액이 더스트 한계보다 작으면 수수료에 추가
		fee += changeAmount
	}

	// 수수료가 입력 금액의 50%를 초과하지 않도록 합니다.
	if fee > totalInput/2 {
		return nil, fmt.Errorf("fee is too high: %d satoshis", fee)
	}

	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %v", err)
	}

	// Create UnsignedTransaction
	unsignedTx := &transaction.UnsignedTransaction{
		NetworkID:               network.ID(),
		UnSignedTxEncodedBase64: base64.StdEncoding.EncodeToString(buf.Bytes()),
		Extra: map[string]interface{}{
			"from":   btcReq.From,
			"to":     btcReq.To,
			"amount": btcReq.Amount,
			"fee":    fee,
		},
	}

	return unsignedTx, nil
}

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
	fmt.Printf("Unspent Txs API Response: %s\n", string(body))

	var utxos []UTXO
	err = json.Unmarshal(body, &utxos)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal UTXOs: %v", err)
	}

	return utxos, nil
}

// func combineSignatureWithTransaction(tx *wire.MsgTx, r, s *big.Int, pubKey *btcec.PublicKey) error {
// 	for i := range tx.TxIn {
// 		signature := ecdsa.Sign(pubKey, r, s)
// 		sigScript, err := txscript.NewScriptBuilder().
// 			AddData(signature.Serialize()).
// 			AddData(pubKey.SerializeCompressed()).
// 			Script()
// 		if err != nil {
// 			return fmt.Errorf("failed to create signature script: %v", err)
// 		}
// 		tx.TxIn[i].SignatureScript = sigScript
// 	}
// 	return nil
// }

func IsValidBitcoinAddress(address string, network Network) bool {
	var params *chaincfg.Params
	switch network {
	case Bitcoin:
		params = &chaincfg.MainNetParams
	case BitcoinTestNet:
		params = &chaincfg.TestNet3Params
	case BitcoinRegTest:
		params = &chaincfg.RegressionNetParams
	default:
		return false
	}

	// 주소 디코딩 시도
	addr, err := btcutil.DecodeAddress(address, params)
	if err != nil {
		return false
	}

	// 주소가 지정된 네트워크에 대해 유효한지 확인
	if !addr.IsForNet(params) {
		return false
	}

	// 주소 유형에 따른 추가 검증
	switch addr.(type) {
	case *btcutil.AddressPubKeyHash:
		return true
	case *btcutil.AddressScriptHash:
		return true
	case *btcutil.AddressWitnessPubKeyHash:
		return true
	case *btcutil.AddressWitnessScriptHash:
		return true
	default:
		// 지원되지 않는 주소 유형
		return false
	}
}
