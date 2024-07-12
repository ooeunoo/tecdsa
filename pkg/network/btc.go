package network

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

const (
	/*
		AddressTypeP2PKH (Pay-to-Public-Key-Hash):
		설명: 가장 전통적인 비트코인 주소 유형입니다.
		특징:
			주소가 '1'로 시작합니다.
			공개키의 해시를 사용합니다.
			레거시 지갑과 호환성이 좋습니다.
	*/
	AddressTypeP2PKH = 0
	/*
		AddressTypeP2SHP2WPKH (Pay-to-Script-Hash wrapping a Pay-to-Witness-Public-Key-Hash):
		설명: 이는 SegWit 주소를 P2SH 형식으로 감싼 것입니다. 흔히 "nested SegWit" 주소라고 불립니다.
		특징:
			주소가 '3'으로 시작합니다.
			SegWit의 이점을 제공하면서도 이전 지갑과의 호환성을 유지합니다.
	*/
	AddressTypeP2SHP2WPKH = 1
	/*
	   AddressTypeP2WPKH (Pay-to-Witness-Public-Key-Hash):
	   설명: 이는 네이티브 SegWit 주소 유형입니다.
	   특징:
	   		주소가 'bc1'로 시작합니다 (메인넷의 경우).
	   		가장 효율적인 트랜잭션 구조를 제공합니다.
	   		더 낮은 트랜잭션 수수료를 가능하게 합니다.
	*/
	AddressTypeP2WPKH = 2
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
		addrType = AddressTypeP2WPKH // 메인넷
	case BitcoinTestNet:
		params = &chaincfg.TestNet3Params
		addrType = AddressTypeP2WPKH // 테스트넷
	default:
		return "", fmt.Errorf("unsupported Bitcoin network: %v", network)
	}

	var address btcutil.Address
	var err error

	switch addrType {
	case AddressTypeP2PKH:
		hash160 := btcutil.Hash160(pubKeyBytes)
		address, err = btcutil.NewAddressPubKeyHash(hash160, params)
	case AddressTypeP2SHP2WPKH:
		witnessProg := btcutil.Hash160(pubKeyBytes)
		witnessAddress, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, params)
		if err != nil {
			return "", fmt.Errorf("failed to create witness address for P2SH-P2WPKH: %w", err)
		}
		address, err = btcutil.NewAddressScriptHash(witnessAddress.ScriptAddress(), params)
	case AddressTypeP2WPKH:
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
