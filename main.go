package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	// 개인키 파일의 전체 경로 생성
	privateKeyPath := filepath.Join(usr.HomeDir, ".ssh", "private_key.pem")

	// 개인키 로드
	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// 공개키 추출
	publicKey := &privateKey.PublicKey

	// 서명할 메시지
	message := []byte("Hello, World!")

	// 메시지 서명
	signature, err := signMessage(privateKey, message)
	if err != nil {
		log.Fatalf("Failed to sign message: %v", err)
	}

	fmt.Printf("Signature: %x\n", signature)

	// 서명 검증
	err = verifySignature(publicKey, message, signature)
	if err != nil {
		log.Fatalf("Signature verification failed: %v", err)
	}

	fmt.Println("Signature verified successfully!")
}
func loadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	keyData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}

	var privateKey *rsa.PrivateKey

	switch block.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	default:
		return nil, fmt.Errorf("unsupported key type %q", block.Type)
	}

	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func signMessage(privateKey *rsa.PrivateKey, message []byte) ([]byte, error) {
	hashed := sha256.Sum256(message)
	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
}

func verifySignature(publicKey *rsa.PublicKey, message, signature []byte) error {
	hashed := sha256.Sum256(message)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
}
