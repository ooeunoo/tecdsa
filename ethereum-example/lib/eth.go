package lib

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

func InjectTestEther(client *ethclient.Client, privateKey string, toAddress string, amount *big.Int) (*types.Receipt, error) {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatalf("failed to load private key: %v", err)
	}

	signedTx, err := SignTransactionWithPrivateKey(client, pk, common.HexToAddress(toAddress), amount)
	if err != nil {
		log.Fatalf("failed to sign transaction: %v", err)
	}

	receipt, err := SendSignedTransaction(client, signedTx, true)
	if err != nil {
		log.Fatalf("failed to send signed transaction: %v", err)
	}

	return receipt, nil

}

// 로우 트랜잭션 생성
func CreateUnsignedTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *types.Transaction {
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    amount,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})
	return tx
}

// Private Key를 사용하여 트랜잭션 생성 및 서명 -> 서명 트랜잭션
func SignTransactionWithPrivateKey(client *ethclient.Client, privateKey *ecdsa.PrivateKey, toAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(15))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(10))

	tx := CreateUnsignedTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	return signedTx, nil
}

// 서명 트랜잭션 브로드 케스트
func SendSignedTransaction(client *ethclient.Client, signedTx *types.Transaction, wait ...bool) (*types.Receipt, error) {
	// wait의 기본값을 false로 설정
	shouldWait := false
	if len(wait) > 0 {
		shouldWait = wait[0]
	}

	// 트랜잭션 전송
	err := client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	// 트랜잭션 완료를 기다릴지 결정
	if shouldWait {
		receipt, err := waitForTransaction(client, signedTx.Hash())
		if err != nil {
			return nil, fmt.Errorf("error waiting for transaction: %v", err)
		}
		return receipt, nil
	}

	return nil, nil
}

// 트랜잭션 완료를 기다리는 헬퍼 함수
func waitForTransaction(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			if err == ethereum.NotFound {
				time.Sleep(time.Second) // 1초 대기 후 다시 시도
				continue
			}
			return nil, err
		}
		return receipt, nil
	}
}

func GenerateRlpEncodedTx(client ethclient.Client, signer types.Signer, fromAddress common.Address, toAddress common.Address, amount *big.Int) (*types.Transaction, []byte) {
	// sign
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("failed to get nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("failed to get gas price: %v", err)
	}
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(15))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(10))

	gasLimit := uint64(21000)
	tx := CreateUnsignedTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)
	rlpEncodedTx, _ := rlp.EncodeToBytes([]interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.To(),
		tx.Value(),
		tx.Data(),
		signer.ChainID(), uint(0), uint(0),
	})
	return tx, rlpEncodedTx
}

func CombineETHUnsignedTxWithSignature(tx *types.Transaction, chainId *big.Int, response SignResponse) (*types.Transaction, string, error) {
	v := int(response.V)
	rBytes, err := base64.StdEncoding.DecodeString(response.R)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode R: %v", err)
	}
	sBytes, err := base64.StdEncoding.DecodeString(response.S)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode S: %v", err)
	}

	signer := types.NewEIP155Signer(chainId)
	signedTx, _ := tx.WithSignature(signer, append(rBytes, append(sBytes, byte(v))...))
	rawTxBytes, _ := rlp.EncodeToBytes(signedTx)
	signedRawTransaction := hex.EncodeToString(rawTxBytes)
	fmt.Println("signedRawTransaction: ", signedRawTransaction)

	return signedTx, signedRawTransaction, nil
}
