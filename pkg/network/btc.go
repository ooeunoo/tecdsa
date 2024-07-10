package network

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
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

func deriveBitcoinAddress(point curves.Point) (string, error) {
	// 1. 공개키를 바이트 배열로 변환
	pubKeyBytes := point.ToAffineCompressed()
	if len(pubKeyBytes) == 0 {
		return "", fmt.Errorf("failed to convert public key to bytes")
	}

	// 2. btcec 라이브러리의 PublicKey로 변환
	_, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}

	// 3. 네트워크 파라미터 설정 (여기서는 메인넷 사용)
	params := &chaincfg.MainNetParams

	var address btcutil.Address

	switch 2 {
	case AddressTypeP2PKH:
		hash160 := btcutil.Hash160(pubKeyBytes)
		address, err = btcutil.NewAddressPubKeyHash(hash160, params)
		if err != nil {
			return "", fmt.Errorf("failed to create P2PKH address: %w", err)
		}
	case AddressTypeP2SHP2WPKH:
		witnessProg := btcutil.Hash160(pubKeyBytes)
		witnessAddress, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, params)
		if err != nil {
			return "", fmt.Errorf("failed to create witness address for P2SH-P2WPKH: %w", err)
		}
		address, err = btcutil.NewAddressScriptHash(witnessAddress.ScriptAddress(), params)
		if err != nil {
			return "", fmt.Errorf("failed to create P2SH-P2WPKH address: %w", err)
		}
	case AddressTypeP2WPKH:
		witnessProg := btcutil.Hash160(pubKeyBytes)
		address, err = btcutil.NewAddressWitnessPubKeyHash(witnessProg, params)
		if err != nil {
			return "", fmt.Errorf("failed to create P2WPKH address: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported address type: %d", "// TODO: Address type")
	}

	encodedAddress := address.EncodeAddress()
	if encodedAddress == "" {
		return "", fmt.Errorf("failed to encode address")
	}

	return encodedAddress, nil
}
