package database

import (
	codec "tecdsa/internal/codec/dkg"
	"tecdsa/internal/dkls/dkg"
	"tecdsa/internal/utils"

	"log"

	"github.com/pkg/errors"
)

func StoreSecretShare(address string, aliceOutput *dkg.AliceOutput, secret []byte) error {
	log.Printf("Storing secret share for address: %s", address)

	encodedAlice, err := codec.EncodeAliceOutput(aliceOutput)
	if err != nil {
		log.Printf("Failed to encode AliceOutput: %v", err)
		return errors.Wrap(err, "failed to encode AliceOutput")
	}
	log.Printf("Successfully encoded AliceOutput")

	encryptedSecret, err := utils.Encrypt(encodedAlice, secret)
	if err != nil {
		log.Printf("Failed to encrypt AliceOutput: %v", err)
		return errors.Wrap(err, "failed to encrypt AliceOutput")
	}
	log.Printf("Successfully encrypted AliceOutput")

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

func GetSecretShare(address string, secret []byte) (*dkg.AliceOutput, error) {
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

	aliceOutput, err := codec.DecodeAliceOutput(decryptedData)
	if err != nil {
		log.Printf("Failed to decode AliceOutput: %v", err)
		return nil, errors.Wrap(err, "failed to decode AliceOutput")
	}
	log.Printf("Successfully decoded AliceOutput")

	log.Printf("Successfully retrieved and decrypted secret share for address: %s", address)
	return aliceOutput, nil
}
