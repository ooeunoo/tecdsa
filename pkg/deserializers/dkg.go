package deserializer

import (
	"bytes"
	"encoding/gob"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/ot/base/simplest"
	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/dkg"
	"github.com/coinbase/kryptology/pkg/zkp/schnorr"
	"github.com/pkg/errors"
)

func registerTypes() {
	gob.Register(&curves.ScalarK256{})
	gob.Register(&curves.PointK256{})
	gob.Register(&curves.ScalarP256{})
	gob.Register(&curves.PointP256{})
}

func EncodeAliceDkgOutput(result *dkg.AliceOutput) ([]byte, error) {
	registerTypes()
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(result); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func EncodeBobDkgOutput(result *dkg.BobOutput) ([]byte, error) {
	registerTypes()
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(result); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeAliceDkgResult(m []byte) (*dkg.AliceOutput, error) {
	registerTypes()
	buf := bytes.NewBuffer(m)
	dec := gob.NewDecoder(buf)
	decoded := new(dkg.AliceOutput)
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func DecodeBobDkgResult(m []byte) (*dkg.BobOutput, error) {
	registerTypes()

	buf := bytes.NewBuffer(m)
	dec := gob.NewDecoder(buf)
	decoded := new(dkg.BobOutput)
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound1Output(commitment [32]byte) ([]byte, error) {
	registerTypes()
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&commitment); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeDkgRound2Input(payload []byte) ([32]byte, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := [32]byte{}
	if err := dec.Decode(&decoded); err != nil {
		return [32]byte{}, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound2Output(output *dkg.Round2Output) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(output); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeDkgRound3Input(payload []byte) (*dkg.Round2Output, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := new(dkg.Round2Output)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound3Output(proof *schnorr.Proof) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(proof); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeDkgRound4Input(payload []byte) (*schnorr.Proof, error) {
	registerTypes()
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := new(schnorr.Proof)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound4Output(proof *schnorr.Proof) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(proof); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeDkgRound5Input(payload []byte) (*schnorr.Proof, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := new(schnorr.Proof)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound5Output(proof *schnorr.Proof) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(proof); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeDkgRound6Input(payload []byte) (*schnorr.Proof, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := new(schnorr.Proof)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound6Output(choices []simplest.ReceiversMaskedChoices) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(choices); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeDkgRound7Input(payload []byte) ([]simplest.ReceiversMaskedChoices, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := []simplest.ReceiversMaskedChoices{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound7Output(challenge []simplest.OtChallenge) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(challenge); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeDkgRound8Input(payload []byte) ([]simplest.OtChallenge, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := []simplest.OtChallenge{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound8Output(responses []simplest.OtChallengeResponse) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(responses); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeDkgRound9Input(payload []byte) ([]simplest.OtChallengeResponse, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := []simplest.OtChallengeResponse{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound9Output(opening []simplest.ChallengeOpening) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(opening); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeDkgRound10Input(payload []byte) ([]simplest.ChallengeOpening, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := []simplest.ChallengeOpening{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}
