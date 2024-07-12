package service

import (
	"fmt"
	"tecdsa/pkg/network"
	"tecdsa/proto/transaction"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

type AddressDerivationFunc func(curves.Point, network.Network) (string, error)
type TransactionHandlerFunc func(*transaction.Transaction) (*transaction.Transaction, error)
type SignatureVerifierFunc func(curves.Point, []byte, []byte) bool

type NetworkHandler struct {
	AddressDerivation  AddressDerivationFunc
	SignatureVerifier  SignatureVerifierFunc
	TransactionHandler TransactionHandlerFunc
}

type NetworkService struct {
	networkHandlerMap map[network.Network]NetworkHandler
}

func NewNetworkService() *NetworkService {
	return &NetworkService{
		networkHandlerMap: map[network.Network]NetworkHandler{
			network.Bitcoin: {
				AddressDerivation: network.DeriveBitcoinAddress,
				// SignatureVerifier:  verifyBitcoinSignature,
				// TransactionHandler: handleBitcoinTransaction,
			},
			network.BitcoinTestNet: {
				AddressDerivation: network.DeriveBitcoinAddress,
				// SignatureVerifier:  verifyBitcoinTestNetSignature,
				// TransactionHandler: handleBitcoinTestNetTransaction,
			},
			network.Ethereum: {
				AddressDerivation: network.DeriveEthereumAddress,
				SignatureVerifier: network.VerifyEtherumSignature,
				// TransactionHandler: handleEthereumTransaction,
			},
			network.Ethereum_Sepolia: {
				AddressDerivation: network.DeriveEthereumAddress,
				SignatureVerifier: network.VerifyEtherumSignature,
				// TransactionHandler: handleEthereumSepoliaTransaction,
			},
		},
	}
}

func (s *NetworkService) GetNetworkByID(id int32) (network.Network, error) {
	switch id {
	case 1:
		return network.Bitcoin, nil
	case 2:
		return network.BitcoinTestNet, nil
	case 3:
		return network.Ethereum, nil
	case 4:
		return network.Ethereum_Sepolia, nil
	default:
		return 0, fmt.Errorf("unsupported network ID: %d", id)
	}
}

func (s *NetworkService) DeriveAddress(point curves.Point, network network.Network) (string, error) {
	handler, exists := s.networkHandlerMap[network]
	if !exists {
		return "", fmt.Errorf("unsupported network: %s", network)
	}
	return handler.AddressDerivation(point, network)
}

func (s *NetworkService) VerifySignature(point curves.Point, network network.Network, txOrigin []byte, signature []byte) (bool, error) {
	handler, exists := s.networkHandlerMap[network]
	if !exists {
		return false, fmt.Errorf("unsupported network: %s", network)
	}
	return handler.SignatureVerifier(point, txOrigin, signature), nil
}

func (s *NetworkService) HandleTransaction(tx *transaction.Transaction, network network.Network) (*transaction.Transaction, error) {
	handler, exists := s.networkHandlerMap[network]
	if !exists {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
	return handler.TransactionHandler(tx)
}

func (s *NetworkService) GetAllNetworks() []network.Network {
	return network.Networks
}

// // 각 네트워크별 구체적인 함수 구현
// func deriveBitcoinAddress(point curves.Point, network Network) (string, error) {
// 	// Bitcoin 주소 도출 로직 구현
// 	return "bitcoin_address", nil
// }

// func verifyBitcoinSignature(point curves.Point, txOrigin []byte, signature []byte) bool {
// 	// Bitcoin 서명 검증 로직 구현
// 	return true
// }

// func handleBitcoinTransaction(tx *transaction.Transaction) (*transaction.Transaction, error) {
// 	// Bitcoin 트랜잭션 처리 로직 구현
// 	return tx, nil
// }

// func deriveBitcoinTestNetAddress(point curves.Point, network Network) (string, error) {
// 	// Bitcoin TestNet 주소 도출 로직 구현
// 	return "bitcoin_testnet_address", nil
// }

// func verifyBitcoinTestNetSignature(point curves.Point, txOrigin []byte, signature []byte) bool {
// 	// Bitcoin TestNet 서명 검증 로직 구현
// 	return true
// }

// func handleBitcoinTestNetTransaction(tx *transaction.Transaction) (*transaction.Transaction, error) {
// 	// Bitcoin TestNet 트랜잭션 처리 로직 구현
// 	return tx, nil
// }

// func deriveEthereumAddress(point curves.Point, network Network) (string, error) {
// 	// Ethereum 주소 도출 로직 구현
// 	return "ethereum_address", nil
// }

// func verifyEthereumSignature(point curves.Point, txOrigin []byte, signature []byte) bool {
// 	// Ethereum 서명 검증 로직 구현
// 	return true
// }

// func handleEthereumTransaction(tx *transaction.Transaction) (*transaction.Transaction, error) {
// 	// Ethereum 트랜잭션 처리 로직 구현
// 	return tx, nil
// }

// func deriveEthereumSepoliaAddress(point curves.Point, network Network) (string, error) {
// 	// Ethereum Sepolia 주소 도출 로직 구현
// 	return "ethereum_sepolia_address", nil
// }

// func verifyEthereumSepoliaSignature(point curves.Point, txOrigin []byte, signature []byte) bool {
// 	// Ethereum Sepolia 서명 검증 로직 구현
// 	return true
// }

// func handleEthereumSepoliaTransaction(tx *transaction.Transaction) (*transaction.Transaction, error) {
// 	// Ethereum Sepolia 트랜잭션 처리 로직 구현
// 	return tx, nil
// }
