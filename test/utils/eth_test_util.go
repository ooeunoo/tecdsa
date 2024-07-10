package test_util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func GenerateETHTxOrigin(toAddress string) (*types.Transaction, []byte, error) {
	// 테스트용 값 설정
	nonce := uint64(0)
	to := common.HexToAddress(toAddress)
	amount := big.NewInt(1000000000000000000) // 1 ETH
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(20000000000) // 20 Gwei
	data := []byte{}
	chainID := big.NewInt(1) // Ethereum Mainnet

	// 트랜잭션 생성
	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)

	// RLP 인코딩
	txHash, err := rlp.EncodeToBytes([]interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.To(),
		tx.Value(),
		tx.Data(),
		chainID, uint(0), uint(0),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to RLP encode transaction: %v", err)
	}

	return tx, txHash, nil
}

func CombineETHUnsignedTxWithSignature(tx *types.Transaction, response []byte) (*types.Transaction, error) {
	var signResponse SignResponse
	err := json.Unmarshal(response, &signResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal sign response: %v", err)
	}

	// V를 int64로 변환
	v, err := strconv.ParseInt(signResponse.V, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'v' as int64: %v", err)
	}

	// R과 S를 big.Int로 변환
	rBytes, err := base64.StdEncoding.DecodeString(signResponse.R)
	if err != nil {
		return nil, fmt.Errorf("failed to decode 'r': %v", err)
	}
	r := new(big.Int).SetBytes(rBytes)

	sBytes, err := base64.StdEncoding.DecodeString(signResponse.S)
	if err != nil {
		return nil, fmt.Errorf("failed to decode 's': %v", err)
	}
	s := new(big.Int).SetBytes(sBytes)

	// 서명 생성 (r + s + v)
	signatureBytes := append(r.Bytes(), s.Bytes()...)
	signatureBytes = append(signatureBytes, byte(v))

	// 로깅 (필요한 경우)
	// fmt.Printf("Generated Signature: %x\n", signatureBytes)
	// fmt.Printf("Signature length: %d\n", len(signatureBytes))

	signer := types.NewEIP155Signer(big.NewInt(1)) // 체인 ID를 1로 가정 (메인넷)
	txWithSignature, err := tx.WithSignature(signer, signatureBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to apply signature to transaction: %v", err)
	}

	return txWithSignature, nil
}

func PrintETHSignedTxAsJSON(signedTx *types.Transaction) {
	signer := types.LatestSignerForChainID(signedTx.ChainId())
	from, _ := types.Sender(signer, signedTx)

	v, r, s := signedTx.RawSignatureValues()

	txJSON := struct {
		Nonce    uint64          `json:"nonce"`
		GasPrice *big.Int        `json:"gasPrice"`
		GasLimit uint64          `json:"gasLimit"`
		To       *common.Address `json:"to"`
		Value    *big.Int        `json:"value"`
		Data     hexutil.Bytes   `json:"data"`
		From     common.Address  `json:"from"`
		V        *big.Int        `json:"v"`
		R        *big.Int        `json:"r"`
		S        *big.Int        `json:"s"`
	}{
		Nonce:    signedTx.Nonce(),
		GasPrice: signedTx.GasPrice(),
		GasLimit: signedTx.Gas(),
		To:       signedTx.To(),
		Value:    signedTx.Value(),
		Data:     signedTx.Data(),
		From:     from,
		V:        v,
		R:        r,
		S:        s,
	}

	jsonData, err := json.MarshalIndent(txJSON, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal transaction to JSON: %v", err)
	}
	fmt.Printf("Signed Transaction as JSON:\n%s\n", string(jsonData))
}


func VerifyETHSignature() {
	
}