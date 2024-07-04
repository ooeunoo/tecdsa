package codec

import (
	"fmt"
)

const (
	DigestSize = 32
)

type Seed [DigestSize]byte
type Commitment = []byte

func MarshalProtoSeed(s Seed) ([]byte, error) {
	return s[:], nil
}

func UnmarshalProtoSeed(data []byte) (Seed, error) {
	if len(data) != DigestSize {
		return Seed{}, fmt.Errorf("invalid seed length")
	}
	var s Seed
	copy(s[:], data)
	return s, nil
}

func MarshalProtoCommitment(c Commitment) ([]byte, error) {
	return c, nil
}

func UnmarshalProtoCommitment(data []byte) (Commitment, error) {
	c := make(Commitment, len(data))
	copy(c, data)
	return c, nil
}
