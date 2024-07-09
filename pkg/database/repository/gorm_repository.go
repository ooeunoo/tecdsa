package repository

import (
	"tecdsa/pkg/database/models"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type GormSecretRepository struct {
	db *gorm.DB
}

func NewGormSecretRepository(db *gorm.DB) *GormSecretRepository {
	return &GormSecretRepository{db: db}
}

func (r *GormSecretRepository) StoreSecretShare(address string, share []byte, secret []byte) error {
	secretRecord := models.Secret{
		Address:   address,
		Share:     share,
		SecretKey: secret,
	}

	if err := r.db.Create(&secretRecord).Error; err != nil {
		return errors.Wrap(err, "failed to store secret in database")
	}

	return nil
}

func (r *GormSecretRepository) GetSecretShare(address string, secret []byte) ([]byte, error) {
	var record models.Secret
	if err := r.db.Where("address = ?", address).First(&record).Error; err != nil {
		return nil, errors.Wrap(err, "failed to retrieve secret from database")
	}
	return record.Share, nil
}
