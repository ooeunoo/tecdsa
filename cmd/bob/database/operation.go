package database

import (
	codec "tecdsa/internal/codec/dkg"
	"tecdsa/internal/dkls/dkg"
	"tecdsa/internal/utils"

	"log"

	"github.com/pkg/errors"
)

func StoreSecretShare(address string, bobOutput *dkg.BobOutput, secret []byte) error {
	log.Printf("Storing secret share for address: %s", address)

	encodedBob, err := codec.EncodeBobOutput(bobOutput)
	if err != nil {
		log.Printf("Failed to encode BobOutput: %v", err)
		return errors.Wrap(err, "failed to encode BobOutput")
	}
	log.Printf("Successfully encoded BobOutput")

	encryptedSecret, err := utils.Encrypt(encodedBob, secret)
	if err != nil {
		log.Printf("Failed to encrypt BobOutput: %v", err)
		return errors.Wrap(err, "failed to encrypt BobOutput")
	}
	log.Printf("Successfully encrypted BobOutput")

	secretRecord := Secret{
		Address:         address,
		EncryptedSecret: encryptedSecret,
		SecretKey:       secret,
	}

	if err := DB.Create(&secretRecord).Error; err != nil {
		log.Printf("Failed to store secret in database: %v", err)
		return errors.Wrap(err, "failed to store secret in database")
	}
	log.Printf("Successfully stored secret share in database for address: %s", address)

	return nil
}

func GetSecretShare(address string, secret []byte) (*dkg.BobOutput, error) {
	log.Printf("Retrieving secret share for address: %s", address)

	var secretRecord Secret
	if err := DB.Where("address = ?", address).First(&secretRecord).Error; err != nil {
		log.Printf("Failed to retrieve secret from database: %v", err)
		return nil, errors.Wrap(err, "failed to retrieve secret from database")
	}
	log.Printf("Successfully retrieved secret record from database")

	// Verify that the provided secret matches the stored secret
	if !utils.SecureCompare(secret, secretRecord.SecretKey) {
		log.Printf("Provided secret does not match stored secret for address: %s", address)
		return nil, errors.New("provided secret does not match stored secret")
	}
	log.Printf("Secret verification successful")

	decryptedData, err := utils.Decrypt(secretRecord.EncryptedSecret, secret)
	if err != nil {
		log.Printf("Failed to decrypt secret: %v", err)
		return nil, errors.Wrap(err, "failed to decrypt secret")
	}
	log.Printf("Successfully decrypted secret data")

	bobOutput, err := codec.DecodeBobOutput(decryptedData)
	if err != nil {
		log.Printf("Failed to decode BobOutput: %v", err)
		return nil, errors.Wrap(err, "failed to decode BobOutput")
	}
	log.Printf("Successfully decoded BobOutput")

	log.Printf("Successfully retrieved and decrypted secret share for address: %s", address)
	return bobOutput, nil
}
