package network

import (
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func deriveEthereumAddress(point curves.Point) (string, error) {
	pointToBytes := point.ToAffineUncompressed()
	unmarshalPubKey, err := crypto.UnmarshalPubkey(pointToBytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal public key")
	}
	address := crypto.PubkeyToAddress(*unmarshalPubKey).Hex()
	return address, nil
}
