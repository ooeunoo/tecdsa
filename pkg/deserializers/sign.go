package deserializer

import (
	"bytes"
	"encoding/gob"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/sign"
	"github.com/pkg/errors"
)

func EncodeSignRound1Output(commitment [32]byte) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&commitment); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeSignRound2Input(payload []byte) ([32]byte, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := [32]byte{}
	if err := dec.Decode(&decoded); err != nil {
		return [32]byte{}, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeSignRound2Output(output *sign.SignRound2Output) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(output); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeSignRound3Input(payload []byte) (*sign.SignRound2Output, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := &sign.SignRound2Output{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeSignRound3Output(output *sign.SignRound3Output) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(output); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeSignRound4Input(payload []byte) (*sign.SignRound3Output, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := &sign.SignRound3Output{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeSignature(signature *curves.EcdsaSignature) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(signature); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeSignature(payload []byte) (*curves.EcdsaSignature, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := &curves.EcdsaSignature{}
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}
