package repository

import "tecdsa/pkg/dkls/dkg"

type SecretRepository interface {
	StoreSecretShare(address string, output dkg.Output, secret []byte) error
	GetSecretShare(address string, secret []byte) (dkg.Output, error)
}
