package network

import (
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

// // 로우 트랜잭션 생성
// func GenerateTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *types.Transaction {
// 	tx := types.NewTx(&types.LegacyTx{
// 		Nonce:    nonce,
// 		To:       &to,
// 		Value:    amount,
// 		Gas:      gasLimit,
// 		GasPrice: gasPrice,
// 		Data:     data,
// 	})
// 	return tx
// }

// // Private Key를 사용하여 트랜잭션 생성 및 서명 -> 서명 트랜잭션
// func SignTransactionWithPrivateKey(client *ethclient.Client, privateKey *ecdsa.PrivateKey, toAddress common.Address, amount *big.Int) (*types.Transaction, error) {
// 	publicKey := privateKey.Public()
// 	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
// 	if !ok {
// 		return nil, fmt.Errorf("error casting public key to ECDSA")
// 	}

// 	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

// 	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get nonce: %v", err)
// 	}

// 	gasLimit := uint64(21000)
// 	gasPrice, err := client.SuggestGasPrice(context.Background())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get gas price: %v", err)
// 	}
// 	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(15))
// 	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(10))

// 	fmt.Println("fromAddress: ", fromAddress.Hex())
// 	fmt.Println("nonce: ", nonce)

// 	tx := GenerateTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)

// 	chainID, err := client.NetworkID(context.Background())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get chain ID: %v", err)
// 	}

// 	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to sign transaction: %v", err)
// 	}

// 	return signedTx, nil
// }

// // 서명 트랜잭션 브로드 케스트
// func SendSignedTransaction(client *ethclient.Client, signedTx *types.Transaction, wait ...bool) error {
// 	// wait의 기본값을 false로 설정
// 	shouldWait := false
// 	if len(wait) > 0 {
// 		shouldWait = wait[0]
// 	}

// 	// 트랜잭션 전송
// 	err := client.SendTransaction(context.Background(), signedTx)
// 	if err != nil {
// 		return fmt.Errorf("failed to send transaction: %v", err)
// 	}

// 	fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())

// 	// 트랜잭션 완료를 기다릴지 결정
// 	if shouldWait {
// 		receipt, err := waitForTransaction(client, signedTx.Hash())
// 		if err != nil {
// 			return fmt.Errorf("error waiting for transaction: %v", err)
// 		}
// 		fmt.Printf("Transaction confirmed in block %d\n", receipt.BlockNumber.Uint64())
// 	}

// 	return nil
// }

// // 트랜잭션 완료를 기다리는 헬퍼 함수
// func waitForTransaction(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
// 	for {
// 		receipt, err := client.TransactionReceipt(context.Background(), txHash)
// 		if err != nil {
// 			if err == ethereum.NotFound {
// 				time.Sleep(time.Second) // 1초 대기 후 다시 시도
// 				continue
// 			}
// 			return nil, err
// 		}
// 		return receipt, nil
// 	}
// }

// // 트랜잭션 RLP 인코딩
// func EncodeTransactionRLP(tx *types.Transaction) ([]byte, error) {
// 	var buf bytes.Buffer
// 	err := tx.EncodeRLP(&buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }

func deriveEthereumAddress(point curves.Point) (string, error) {
	pointToBytes := point.ToAffineUncompressed()
	unmarshalPubKey, err := crypto.UnmarshalPubkey(pointToBytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal public key")
	}
	address := crypto.PubkeyToAddress(*unmarshalPubKey).Hex()
	return address, nil
}
