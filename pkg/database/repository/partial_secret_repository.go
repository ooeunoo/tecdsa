package repository

import (
	"fmt"
	"tecdsa/pkg/database/models"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ParitalSecretShareRepository interface {
	Create(address string, share []byte, clientSecurityID uint) error
	FindByAddress(address string) (*models.ParitalSecretShare, error)
	FindByClientSecurityID(clientSecurityID uint) ([]*models.ParitalSecretShare, error)
}

type paritalSecretShareRepositoryImpl struct {
	db *gorm.DB
}

func NewPartialSecretShareRepository(db *gorm.DB) ParitalSecretShareRepository {
	return &paritalSecretShareRepositoryImpl{db: db}
}

func (r *paritalSecretShareRepositoryImpl) Create(address string, share []byte, clientSecurityID uint) error {
	secretRecord := models.ParitalSecretShare{
		Address:          address,
		Share:            share,
		ClientSecurityID: clientSecurityID,
	}

	if err := r.db.Create(&secretRecord).Error; err != nil {
		fmt.Println("err: ", err)
		return errors.Wrap(err, "failed to store secret in database")
	}

	return nil
}

func (r *paritalSecretShareRepositoryImpl) FindByAddress(address string) (*models.ParitalSecretShare, error) {
	var record models.ParitalSecretShare
	if err := r.db.Where("address = ?", address).First(&record).Error; err != nil {
		return nil, errors.Wrap(err, "failed to retrieve secret from database")
	}
	return &record, nil
}

func (r *paritalSecretShareRepositoryImpl) FindByClientSecurityID(clientSecurityID uint) ([]*models.ParitalSecretShare, error) {
	var records []*models.ParitalSecretShare
	if err := r.db.Where("client_security_id = ?", clientSecurityID).Find(&records).Error; err != nil {
		return nil, errors.Wrap(err, "failed to retrieve secrets from database")
	}
	return records, nil
}
