package deserializer

import (
	"bytes"
	"encoding/gob"

	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/sign"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	pb "tecdsa/proto/sign"
)

func EncodeSignRequestToRound1(req *pb.SignRequestMessage) ([]byte, error) {
	return proto.Marshal(req)
}

func DecodeSignRequestToRound1(payload []byte) (*pb.SignRequestMessage, error) {
	req := &pb.SignRequestMessage{}
	err := proto.Unmarshal(payload, req)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling SignRequest")
	}
	return req, nil
}

func EncodeSignRound1Payload(commitment [32]byte) ([]byte, error) {
	registerTypes()
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&commitment); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeSignRound1Payload(payload []byte) ([32]byte, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := [32]byte{}
	if err := dec.Decode(&decoded); err != nil {
		return [32]byte{}, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeSignRound2Payload(output *sign.SignRound2Output) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(output); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeSignRound2Payload(payload []byte) (*sign.SignRound2Output, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := &sign.SignRound2Output{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}

func EncodeSignRound3Payload(output *sign.SignRound3Output) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(output); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func DecodeSignRound3Payload(payload []byte) (*sign.SignRound3Output, error) {
	buf := bytes.NewBuffer(payload)
	dec := gob.NewDecoder(buf)
	decoded := &sign.SignRound3Output{}
	if err := dec.Decode(&decoded); err != nil {
		return nil, errors.WithStack(err)
	}
	return decoded, nil
}
