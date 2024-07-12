package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"crypto/x509"
	"encoding/pem"

	"github.com/pkg/errors"
)

func GenerateSecretKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate secret key")
	}
	return key, nil
}

func Encrypt(data []byte, secret []byte) ([]byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCM")
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, errors.Wrap(err, "failed to generate nonce")
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func Decrypt(encryptedData []byte, secret []byte) ([]byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCM")
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt")
	}

	return plaintext, nil
}

func SecureCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

// ValidatePublicKey checks if the given public key is valid
func ValidatePublicKey(publicKeyPEM string) error {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	switch pub.(type) {
	case *rsa.PublicKey:
		// For RSA, we might want to check the key size
		rsaKey := pub.(*rsa.PublicKey)
		if rsaKey.N.BitLen() < 2048 {
			return errors.New("RSA key size is less than 2048 bits")
		}
	case *ecdsa.PublicKey:
		// For ECDSA, we might want to check the curve
		ecdsaKey := pub.(*ecdsa.PublicKey)
		curve := ecdsaKey.Curve.Params().Name
		if curve != "P-256" && curve != "P-384" && curve != "P-521" {
			return errors.New("unsupported elliptic curve")
		}
	default:
		return errors.New("unsupported public key type")
	}

	return nil
}
