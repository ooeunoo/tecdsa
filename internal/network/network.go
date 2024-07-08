package network

import (
	"fmt"

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
	// Layer 2 solutions
	Arbitrum
	Optimism
	Loopring
	zkSync
	StarkNet
)

func GetNetworkByID(id int) (Network, error) {
	switch id {
	case 0:
		return Bitcoin, nil
	case 1:
		return Ethereum, nil
	case 2:
		return Litecoin, nil
	case 3:
		return Dogecoin, nil
	case 4:
		return Ripple, nil
	case 5:
		return Cardano, nil
	case 6:
		return Polkadot, nil
	case 7:
		return Solana, nil
	case 8:
		return BinanceSmartChain, nil
	case 9:
		return Polygon, nil
	case 10:
		return Avalanche, nil
	case 11:
		return Tron, nil
	case 12:
		return Cosmos, nil
	case 13:
		return Algorand, nil
	case 14:
		return Stellar, nil
	case 15:
		return Monero, nil
	case 16:
		return Zcash, nil
	case 17:
		return Dash, nil
	case 18:
		return EthereumClassic, nil
	case 19:
		return BitcoinCash, nil
	case 20:
		return Arbitrum, nil
	case 21:
		return Optimism, nil
	case 22:
		return Loopring, nil
	case 23:
		return zkSync, nil
	case 24:
		return StarkNet, nil
	default:
		return 0, fmt.Errorf("unsupported network ID: %d", id)
	}
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

type AddressDerivationFunc func(curves.Point) (string, error)

var addressDerivationMap = map[Network]AddressDerivationFunc{
	// Bitcoin:           deriveBitcoinAddress,
	Ethereum: deriveEthereumAddress,
	// Litecoin:          deriveLitecoinAddress,
	// Dogecoin:          deriveDogecoinAddress,
	// Ripple:            deriveRippleAddress,
	// Cardano:           deriveCardanoAddress,
	// Polkadot:          derivePolkadotAddress,
	// Solana:            deriveSolanaAddress,
	BinanceSmartChain: deriveEthereumAddress,
	Polygon:           deriveEthereumAddress,
	// Avalanche:         deriveAvalancheAddress,
	// Tron:              deriveTronAddress,
	// Cosmos:            deriveCosmosAddress,
	// Algorand:          deriveAlgorandAddress,
	// Stellar:           deriveStellarAddress,
	// Monero:            deriveMoneroAddress,
	// Zcash:             deriveZcashAddress,
	// Dash:              deriveDashAddress,
	EthereumClassic: deriveEthereumAddress,
	// BitcoinCash:       deriveBitcoinCashAddress,
	Arbitrum: deriveEthereumAddress,
	Optimism: deriveEthereumAddress,
	// Loopring:          deriveLoopringAddress,
	// zkSync:            deriveZkSyncAddress,
	// StarkNet:          deriveStarkNetAddress,
}

func DeriveAddress(point curves.Point, network Network) (string, error) {
	derivationFunc, exists := addressDerivationMap[network]
	if !exists {
		return "", fmt.Errorf("unsupported network: %s", network)
	}
	return derivationFunc(point)
}
