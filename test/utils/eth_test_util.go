package eth_test_util

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func GenerateTxOrigin(toAddress string) (*types.Transaction, []byte, error) {
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

func PrintSignedTxAsJSON(signedTx *types.Transaction) {
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
