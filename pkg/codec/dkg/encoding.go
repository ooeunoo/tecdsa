package codec

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"tecdsa/pkg/dkls/dkg"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

func init() {
	gob.Register(&curves.ScalarK256{})
	gob.Register(&curves.PointK256{})
	gob.Register(&dkg.AliceOutput{})
	gob.Register(&dkg.BobOutput{})
}

func EncodeOutput(output dkg.Output) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(output); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeOutput(data []byte) (dkg.Output, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	var output dkg.Output
	if err := dec.Decode(&output); err != nil {
		return nil, err
	}

	switch output.(type) {
	case *dkg.AliceOutput, *dkg.BobOutput:
		return output, nil
	default:
		return nil, fmt.Errorf("unknown output type: %T", output)
	}
}
