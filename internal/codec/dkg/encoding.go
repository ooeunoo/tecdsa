package encoidng

import (
	"bytes"
	"encoding/gob"

	"github.com/pkg/errors"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/ot/base/simplest"
	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/dkg"
	"github.com/coinbase/kryptology/pkg/zkp/schnorr"
)

const payloadKey = "direct"

type DkgRoundMessage struct {
	Payloads map[string][]byte
	Metadata map[string]string
}

func newDkgRoundMessage(payload []byte, round string) *DkgRoundMessage {
	return &DkgRoundMessage{
		Payloads: map[string][]byte{payloadKey: payload},
		Metadata: map[string]string{"round": round},
	}
}

func registerTypes() {
	gob.Register(&curves.ScalarK256{})
	gob.Register(&curves.PointK256{})
	gob.Register(&curves.ScalarP256{})
	gob.Register(&curves.PointP256{})
}

func EncodeDkgRound1Output(commitment [32]byte, version uint) (*DkgRoundMessage, error) {
	registerTypes()
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&commitment); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "1"), nil
}

func DecodeDkgRound2Input(m *DkgRoundMessage) ([32]byte, error) {
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := [32]byte{}
	if err := dec.Decode(&decoded); err != nil {
		return [32]byte{}, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound2Output(output *dkg.Round2Output) (*DkgRoundMessage, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(output); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "2"), nil
}

func DecodeDkgRound3Input(m *DkgRoundMessage) (*dkg.Round2Output, error) {
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(dkg.Round2Output)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound3Output(proof *schnorr.Proof, version uint) (*DkgRoundMessage, error) {

	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(proof); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "3"), nil
}

func DecodeDkgRound4Input(m *DkgRoundMessage) (*schnorr.Proof, error) {
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(schnorr.Proof)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound4Output(proof *schnorr.Proof) (*DkgRoundMessage, error) {

	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(proof); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "4"), nil
}

func DecodeDkgRound5Input(m *DkgRoundMessage) (*schnorr.Proof, error) {
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(schnorr.Proof)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound5Output(proof *schnorr.Proof, version uint) (*DkgRoundMessage, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(proof); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "5"), nil
}

func DecodeDkgRound6Input(m *DkgRoundMessage) (*schnorr.Proof, error) {
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(schnorr.Proof)
	if err := dec.Decode(decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound6Output(choices []simplest.ReceiversMaskedChoices) (*DkgRoundMessage, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(choices); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "6"), nil
}

func DecodeDkgRound7Input(m *DkgRoundMessage) ([]simplest.ReceiversMaskedChoices, error) {
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := []simplest.ReceiversMaskedChoices{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound7Output(challenge []simplest.OtChallenge, version uint) (*DkgRoundMessage, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(challenge); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "7"), nil
}

func DecodeDkgRound8Input(m *DkgRoundMessage) ([]simplest.OtChallenge, error) {
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := []simplest.OtChallenge{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound8Output(responses []simplest.OtChallengeResponse) (*DkgRoundMessage, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(responses); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "8"), nil
}

func DecodeDkgRound9Input(m *DkgRoundMessage) ([]simplest.OtChallengeResponse, error) {
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := []simplest.OtChallengeResponse{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeDkgRound9Output(opening []simplest.ChallengeOpening) (*DkgRoundMessage, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(opening); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "9"), nil
}

func DecodeDkgRound10Input(m *DkgRoundMessage) ([]simplest.ChallengeOpening, error) {
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := []simplest.ChallengeOpening{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

// EncodeAliceDkgOutput serializes Alice DKG
func EncodeAliceDkgOutput(result *dkg.AliceOutput) (*DkgRoundMessage, error) {
	registerTypes()
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(result); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "alice-output"), nil
}

// DecodeAliceDkgResult deserializes Alice DKG output.
func DecodeAliceDkgResult(m *DkgRoundMessage) (*dkg.AliceOutput, error) {
	registerTypes()
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(dkg.AliceOutput)
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

// EncodeBobDkgOutput serializes Bob DKG output
func EncodeBobDkgOutput(result *dkg.BobOutput) (*DkgRoundMessage, error) {
	registerTypes()
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(result); err != nil {
		return nil, errors.WithStack(err)
	}
	return newDkgRoundMessage(buf.Bytes(), "bob-output"), nil
}

// DecodeBobDkgResult deserializes Bob DKG output.
func DecodeBobDkgResult(m *DkgRoundMessage) (*dkg.BobOutput, error) {
	buf := bytes.NewBuffer(m.Payloads[payloadKey])
	dec := gob.NewDecoder(buf)
	decoded := new(dkg.BobOutput)
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}
