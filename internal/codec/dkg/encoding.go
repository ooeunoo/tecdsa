package codec

import (
	"bytes"
	"encoding/gob"

	"tecdsa/internal/dkls/dkg"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

// BobOutput을 직렬화하는 함수
func EncodeBobOutput(output *dkg.BobOutput) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	// curves.Point와 curves.Scalar 타입을 위한 gob 등록
	gob.Register(&curves.ScalarK256{})
	gob.Register(&curves.PointK256{})

	if err := enc.Encode(output); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// AliceOutput을 직렬화하는 함수
func EncodeAliceOutput(output *dkg.AliceOutput) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	// curves.Point와 curves.Scalar 타입을 위한 gob 등록
	gob.Register(&curves.ScalarK256{})
	gob.Register(&curves.PointK256{})

	if err := enc.Encode(output); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// 직렬화된 데이터를 BobOutput으로 역직렬화하는 함수
func DecodeBobOutput(data []byte) (*dkg.BobOutput, error) {
	var output dkg.BobOutput
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&output); err != nil {
		return nil, err
	}
	return &output, nil
}

// 직렬화된 데이터를 AliceOutput으로 역직렬화하는 함수
func DecodeAliceOutput(data []byte) (*dkg.AliceOutput, error) {
	var output dkg.AliceOutput
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&output); err != nil {
		return nil, err
	}
	return &output, nil
}
