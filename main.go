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
	"os"
	"path/filepath"
)

func main() {
	// 홈 디렉토리 얻기
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	// 키 로드
	privateKey, err := loadPrivateKey(filepath.Join(homeDir, ".ssh", "private_key.pem"))
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	publicKey, err := loadPublicKey(filepath.Join(homeDir, ".ssh", "public_key.pem"))
	if err != nil {
		log.Fatalf("Failed to load public key: %v", err)
	}

	// 메시지 준비
	message := []byte("Hello, World!")
	fmt.Printf("Original message: %s\n", message)

	// 암호화
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, message, nil)
	if err != nil {
		log.Fatalf("Error encrypting message: %v", err)
	}
	fmt.Printf("Encrypted message: %x\n", ciphertext)

	// 복호화
	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		log.Fatalf("Error decrypting message: %v", err)
	}
	fmt.Printf("Decrypted message: %s\n", decrypted)

	// 서명
	hashed := sha256.Sum256(message)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		log.Fatalf("Error signing message: %v", err)
	}
	fmt.Printf("Signature: %x\n", signature)

	// 검증
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
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

func loadPublicKey(filename string) (*rsa.PublicKey, error) {
	keyData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, fmt.Errorf("unsupported public key type")
	}
}
