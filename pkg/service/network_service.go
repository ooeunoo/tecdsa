package service

import (
	"fmt"
	"tecdsa/pkg/network"
	"tecdsa/pkg/transaction"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

type AddressDerivationFunc func(curves.Point, network.Network) (string, error)
type SignatureVerifierFunc func(curves.Point, []byte, []byte) bool
type CreateUnsignedTxFunc func(interface{}, network.Network) (*transaction.UnsignedTransaction, error)

type NetworkHandler struct {
	AddressDerivation         AddressDerivationFunc
	SignatureVerifier         SignatureVerifierFunc
	CreateUnsignedTransaction CreateUnsignedTxFunc
}

type NetworkService struct {
	networkHandlerMap map[network.Network]NetworkHandler
}

func NewNetworkService() *NetworkService {
	return &NetworkService{
		networkHandlerMap: map[network.Network]NetworkHandler{
			network.Bitcoin: {
				AddressDerivation:         network.DeriveBitcoinAddress,
				CreateUnsignedTransaction: network.CreateUnsignedBitcoinTransaction,
			},
			network.BitcoinTestNet: {
				AddressDerivation:         network.DeriveBitcoinAddress,
				CreateUnsignedTransaction: network.CreateUnsignedBitcoinTransaction,
			},
			network.Ethereum: {
				AddressDerivation:         network.DeriveEthereumAddress,
				SignatureVerifier:         network.VerifyEtherumSignature,
				CreateUnsignedTransaction: network.CreateUnsignedEthereumTransaction,
			},
			network.Ethereum_Sepolia: {
				AddressDerivation:         network.DeriveEthereumAddress,
				SignatureVerifier:         network.VerifyEtherumSignature,
				CreateUnsignedTransaction: network.CreateUnsignedEthereumTransaction,
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

func (s *NetworkService) GetAllNetworks() []network.Network {
	return network.Networks
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

func (s *NetworkService) CreateUnsignedTransaction(network network.Network, txRequest interface{}) (*transaction.UnsignedTransaction, error) {
	handler, exists := s.networkHandlerMap[network]
	if !exists {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
	return handler.CreateUnsignedTransaction(txRequest, network)
}
