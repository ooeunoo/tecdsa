package repository

import (
	"fmt"
	"tecdsa/pkg/database/models"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ClientSecurityRepository interface {
	Create(publicKey string, ip string) (*models.ClientSecurity, error)
	FindByID(id uint) (*models.ClientSecurity, error)
	FindByIP(ip string) (*models.ClientSecurity, error)
}

type clientSecurityRepositoryImpl struct {
	db *gorm.DB
}

func NewClientSecurityRepository(db *gorm.DB) ClientSecurityRepository {
	return &clientSecurityRepositoryImpl{db: db}
}

func (r *clientSecurityRepositoryImpl) Create(publicKey string, ip string) (*models.ClientSecurity, error) {
	record := &models.ClientSecurity{
		PublicKey: publicKey,
		IP:        ip,
	}
	if err := r.db.Create(record).Error; err != nil {
		fmt.Println(err)
		return nil, errors.Wrap(err, "failed to create client security")
	}
	return record, nil
}

func (r *clientSecurityRepositoryImpl) FindByID(id uint) (*models.ClientSecurity, error) {
	var record models.ClientSecurity
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find client security by ID")
	}
	return &record, nil
}

func (r *clientSecurityRepositoryImpl) FindByIP(ip string) (*models.ClientSecurity, error) {
	var record models.ClientSecurity
	if err := r.db.Where("ip = ?", ip).First(&record).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find client security by IP")
	}
	return &record, nil
}
