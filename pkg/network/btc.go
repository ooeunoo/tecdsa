package network

import (
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcutil/bech32"

	"github.com/btcsuite/btcutil/base58"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"golang.org/x/crypto/ripemd160"
)

func deriveSegWitAddress(point curves.Point) (string, error) {
	// 1. 공개키를 바이트 배열로 변환
	pubKeyBytes := point.ToAffineCompressed()

	// 2. SHA256 해시 적용
	sha256Hash := sha256.Sum256(pubKeyBytes)

	// 3. RIPEMD160 해시 적용
	ripemd160Hasher := ripemd160.New()
	_, err := ripemd160Hasher.Write(sha256Hash[:])
	if err != nil {
		return "", err
	}
	pubKeyHash := ripemd160Hasher.Sum(nil)
	// 4. 0 버전과 함께 witness program 생성
	program := append([]byte{0x00}, pubKeyHash...)

	// 5. Convert program to 5-bit words for bech32 encoding
	converted, err := bech32.ConvertBits(program, 8, 5, true)
	if err != nil {
		return "", err
	}

	// 6. Bech32 인코딩 적용
	address, err := bech32.Encode("bc", converted)
	if err != nil {
		return "", err
	}

	fmt.Println("address: ", address)
	return address, nil
}

func deriveBitcoinAddress(point curves.Point) (string, error) {
	// 1. 공개키를 바이트 배열로 변환
	pubKeyBytes := point.ToAffineCompressed()

	// 2. SHA256 해시 적용
	sha256Hash := sha256.Sum256(pubKeyBytes)

	// 3. RIPEMD160 해시 적용
	ripemd160Hasher := ripemd160.New()
	_, err := ripemd160Hasher.Write(sha256Hash[:])
	if err != nil {
		return "", err
	}
	pubKeyHash := ripemd160Hasher.Sum(nil)

	// 4. 버전 바이트 추가 (0x00 for mainnet)
	versionedPayload := append([]byte{0x00}, pubKeyHash...)

	// 5. 체크섬 생성 (두 번의 SHA256)
	firstSHA := sha256.Sum256(versionedPayload)
	secondSHA := sha256.Sum256(firstSHA[:])
	checksum := secondSHA[:4]

	// 6. 체크섬을 주소에 추가
	fullPayload := append(versionedPayload, checksum...)

	// 7. Base58Check 인코딩 적용
	address := base58.Encode(fullPayload)

	return address, nil
}
