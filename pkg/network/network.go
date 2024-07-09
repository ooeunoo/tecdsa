package network

import (
	"fmt"
	"tecdsa/proto/transaction"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

type Network int

const (
	Bitcoin Network = iota
	Ethereum
	Litecoin
	Dogecoin
	Ripple
	Cardano
	Polkadot
	Solana
	BinanceSmartChain
	Polygon
	Avalanche
	Tron
	Cosmos
	Algorand
	Stellar
	Monero
	Zcash
	Dash
	EthereumClassic
	BitcoinCash
	// Layer 2
	Arbitrum
	Optimism
	Loopring
	zkSync
	StarkNet
)

var Networks = []Network{
	Bitcoin, Ethereum, Litecoin, Dogecoin, Ripple,
	Cardano, Polkadot, Solana, BinanceSmartChain, Polygon,
	Avalanche, Tron, Cosmos, Algorand, Stellar,
	Monero, Zcash, Dash, EthereumClassic, BitcoinCash,
	Arbitrum, Optimism, Loopring, zkSync, StarkNet,
}

func (n Network) String() string {
	return [...]string{
		"Bitcoin", "Ethereum", "Litecoin", "Dogecoin", "Ripple",
		"Cardano", "Polkadot", "Solana", "Binance Smart Chain", "Polygon",
		"Avalanche", "Tron", "Cosmos", "Algorand", "Stellar",
		"Monero", "Zcash", "Dash", "Ethereum Classic", "Bitcoin Cash",
		"Arbitrum", "Optimism", "Loopring", "zkSync", "StarkNet",
	}[n]
}
func GetNetworkByID(id int) (Network, error) {
	switch id {
	case 1:
		return Bitcoin, nil
	case 2:
		return Ethereum, nil
	case 3:
		return Litecoin, nil
	case 4:
		return Dogecoin, nil
	case 5:
		return Ripple, nil
	case 6:
		return Cardano, nil
	case 7:
		return Polkadot, nil
	case 8:
		return Solana, nil
	case 9:
		return BinanceSmartChain, nil
	case 10:
		return Polygon, nil
	case 11:
		return Avalanche, nil
	case 12:
		return Tron, nil
	case 13:
		return Cosmos, nil
	case 14:
		return Algorand, nil
	case 15:
		return Stellar, nil
	case 16:
		return Monero, nil
	case 17:
		return Zcash, nil
	case 18:
		return Dash, nil
	case 19:
		return EthereumClassic, nil
	case 20:
		return BitcoinCash, nil
	case 21:
		return Arbitrum, nil
	case 22:
		return Optimism, nil
	case 23:
		return Loopring, nil
	case 24:
		return zkSync, nil
	case 25:
		return StarkNet, nil
	default:
		return 0, fmt.Errorf("unsupported network ID: %d", id)
	}
}

type AddressDerivationFunc func(curves.Point) (string, error)
type TransactionHandlerFunc func(*transaction.Transaction) (*transaction.Transaction, error)

type NetworkHandler struct {
	AddressDerivation  AddressDerivationFunc
	TransactionHandler TransactionHandlerFunc
}

var networkHandlerMap = map[Network]NetworkHandler{
	Bitcoin: {
		// AddressDerivation:  deriveBitcoinAddress,
		// TransactionHandler: handleBitcoinTransaction,
	},
	Ethereum: {
		AddressDerivation: deriveEthereumAddress,
		// TransactionHandler: handleEthereumTransaction,
	},
	BinanceSmartChain: {
		// AddressDerivation:  deriveEthereumAddress,
		// TransactionHandler: handleEthereumTransaction,
	},
	Polygon: {
		// AddressDerivation:  deriveEthereumAddress,
		// TransactionHandler: handleEthereumTransaction,
	},
	EthereumClassic: {
		// AddressDerivation:  deriveEthereumAddress,
		// TransactionHandler: handleEthereumTransaction,
	},
	Arbitrum: {
		// AddressDerivation:  deriveEthereumAddress,
		// TransactionHandler: handleEthereumTransaction,
	},
	Optimism: {
		// AddressDerivation:  deriveEthereumAddress,
		// TransactionHandler: handleEthereumTransaction,
	},
	// 다른 네트워크들에 대한 핸들러 추가...
}

func DeriveAddress(point curves.Point, network Network) (string, error) {
	handler, exists := networkHandlerMap[network]
	if !exists {
		return "", fmt.Errorf("unsupported network: %s", network)
	}
	return handler.AddressDerivation(point)
}

func HandleTransaction(tx *transaction.Transaction, network Network) (*transaction.Transaction, error) {
	handler, exists := networkHandlerMap[network]
	if !exists {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
	return handler.TransactionHandler(tx)
}
