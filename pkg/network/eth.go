package network

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"tecdsa/pkg/transaction"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
)

type EthereumTxRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
	/*
	   	Value
	    wei (1ETH = 1000000000000000000)
	*/
	Amount string  `json:"amount"`
	Nonce  *uint64 `json:"nonce,omitempty"`
	/*
		GasPrice
		단위는 Wei입니다 (1 Gwei = 10^9 Wei)
		네트워크 상황에 따라 적절한 값이 달라집니다
		일반적으로 "20000000000" (20 Gwei) 정도가 적당할 수 있습니다
	*/
	GasPrice *string `json:"gasPrice,omitempty"` // wei (20Gwei= "20000000000")
	/*
		GasLimit
		단순 ETH 전송의 경우 21000이 표준입니다
		스마트 컨트랙트 상호작용의 경우 더 높은 값이 필요합니다
	*/
	GasLimit *uint64 `json:"gasLimit,omitempty"`
	Data     string  `json:"data,omitempty"`
}

func DeriveEthereumAddress(point curves.Point, _ Network) (string, error) {
	pointToBytes := point.ToAffineUncompressed()
	unmarshalPubKey, err := crypto.UnmarshalPubkey(pointToBytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal public key")
	}
	address := crypto.PubkeyToAddress(*unmarshalPubKey).Hex()
	return address, nil
}

func VerifyEtherumSignature(point curves.Point, txOrigin []byte, signature []byte) bool {
	publicKey := point.ToAffineUncompressed()
	return crypto.VerifySignature(publicKey, crypto.Keccak256(txOrigin), signature)
}

func CreateUnsignedEthereumTransaction(req interface{}, network Network) (*transaction.UnsignedTransaction, error) {
	ethReq, ok := req.(EthereumTxRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for Ethereum transaction")
	}
	// Validate addresses
	if !IsValidEthereumAddress(ethReq.From) {
		return nil, fmt.Errorf("invalid 'from' address: %s", ethReq.From)
	}
	if !IsValidEthereumAddress(ethReq.To) {
		return nil, fmt.Errorf("invalid 'to' address: %s", ethReq.To)
	}

	chainID := network.ChainID()
	if chainID == nil {
		return nil, fmt.Errorf("invalid chain ID for network: %s", network)
	}
	chainIDBigInt := new(big.Int).SetInt64(*chainID)

	client, err := ethclient.Dial(network.RPC())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	to := common.HexToAddress(ethReq.To)
	value, ok := new(big.Int).SetString(ethReq.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid value: %s", ethReq.Amount)
	}

	nonce, err := getNonce(client, ethReq.From, ethReq.Nonce)
	if err != nil {
		return nil, err
	}

	gasPrice, err := getGasPrice(client, ethReq.GasPrice)
	if err != nil {
		return nil, err
	}

	data, err := getTransactionData(ethReq.Data)
	if err != nil {
		return nil, err
	}

	gasLimit, err := getGasLimit(client, ethReq.To, data, ethReq.GasLimit)
	if err != nil {
		return nil, err
	}

	// Create transaction
	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)

	// RLP encoding
	txHash, err := rlp.EncodeToBytes([]interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.To(),
		tx.Value(),
		tx.Data(),
		chainIDBigInt, uint(0), uint(0),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to RLP encode transaction: %v", err)
	}

	// Calculate transaction hash
	// hash := tx.Hash().Hex()

	// Create UnsignedTransaction
	unsignedTx := &transaction.UnsignedTransaction{
		NetworkID:               network.ID(),
		UnSignedTxEncodedBase64: base64.StdEncoding.EncodeToString(txHash),
		Extra: map[string]interface{}{
			"nonce":    tx.Nonce(),
			"gasPrice": tx.GasPrice().String(),
			"gasLimit": tx.Gas(),
			"to":       tx.To().Hex(),
			"value":    tx.Value().String(),
			"data":     hex.EncodeToString(tx.Data()),
		},
	}

	return unsignedTx, nil
}

func getNonce(client *ethclient.Client, address string, providedNonce *uint64) (uint64, error) {
	if providedNonce != nil {
		return *providedNonce, nil
	}
	return client.PendingNonceAt(context.Background(), common.HexToAddress(address))
}

func getGasPrice(client *ethclient.Client, providedGasPrice *string) (*big.Int, error) {
	if providedGasPrice != nil {
		gasPrice, ok := new(big.Int).SetString(*providedGasPrice, 10)
		if !ok {
			return nil, fmt.Errorf("invalid gas price: %s", *providedGasPrice)
		}
		return gasPrice, nil
	}
	return client.SuggestGasPrice(context.Background())
}

func getGasLimit(client *ethclient.Client, to string, data []byte, providedGasLimit *uint64) (uint64, error) {
	if providedGasLimit != nil {
		return *providedGasLimit, nil
	}
	toAddress := common.HexToAddress(to)
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &toAddress,
		Data: data,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas limit: %v", err)
	}

	return uint64(float64(gasLimit) * 1.2), nil
}

func getTransactionData(providedData string) ([]byte, error) {
	if providedData == "" {
		return nil, nil
	}
	return hex.DecodeString(providedData)
}
func IsValidEthereumAddress(address string) bool {
	return common.IsHexAddress(address)
}
