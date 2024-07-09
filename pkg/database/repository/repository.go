package repository

type SecretRepository interface {
	StoreSecretShare(address string, share []byte, secret []byte) error
	GetSecretShare(address string, secret []byte) ([]byte, error)
}
