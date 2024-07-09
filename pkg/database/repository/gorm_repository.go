package repository

import (
	codec "tecdsa/pkg/codec/dkg"
	"tecdsa/pkg/database/models"
	"tecdsa/pkg/dkls/dkg"
	"tecdsa/pkg/utils"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type GormSecretRepository struct {
	db *gorm.DB
}

func NewGormSecretRepository(db *gorm.DB) *GormSecretRepository {
	return &GormSecretRepository{db: db}
}

func (r *GormSecretRepository) StoreSecretShare(address string, output dkg.Output, secret []byte) error {
	encodedOutput, err := codec.EncodeOutput(output)
	if err != nil {
		return errors.Wrap(err, "failed to encode Output")
	}

	encryptedSecret, err := utils.Encrypt(encodedOutput, secret)
	if err != nil {
		return errors.Wrap(err, "failed to encrypt Output")
	}

	secretRecord := models.Secret{
		Address:         address,
		EncryptedSecret: encryptedSecret,
		SecretKey:       secret,
	}

	if err := r.db.Create(&secretRecord).Error; err != nil {
		return errors.Wrap(err, "failed to store secret in database")
	}

	return nil
}

func (r *GormSecretRepository) GetSecretShare(address string, secret []byte) (dkg.Output, error) {
	var secretRecord models.Secret
	if err := r.db.Where("address = ?", address).First(&secretRecord).Error; err != nil {
		return nil, errors.Wrap(err, "failed to retrieve secret from database")
	}

	if !utils.SecureCompare(secret, secretRecord.SecretKey) {
		return nil, errors.New("provided secret does not match stored secret")
	}

	decryptedData, err := utils.Decrypt(secretRecord.EncryptedSecret, secret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt secret")
	}

	output, err := codec.DecodeOutput(decryptedData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode Output")
	}

	return output, nil
}
