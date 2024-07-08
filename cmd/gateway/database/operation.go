package database

import "github.com/pkg/errors"

func StoreSecretShare(address string, encryptedSecret []byte) error {
	secret := Secret{
		Address:         address,
		EncryptedSecret: encryptedSecret,
	}

	if err := DB.Create(&secret).Error; err != nil {
		return errors.Wrap(err, "failed to store secret in database")
	}

	return nil
}

func GetSecretShare(address string) ([]byte, error) {
	var secret Secret
	if err := DB.Where("address = ?", address).First(&secret).Error; err != nil {
		return nil, errors.Wrap(err, "failed to retrieve secret from database")
	}

	return secret.EncryptedSecret, nil
}
