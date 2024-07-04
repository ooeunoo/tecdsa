package dkg

import (
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/tecdsa/dkls/v1/dkg"
)

type Alice struct {
	*dkg.Alice
}

func NewAlice(curve *curves.Curve) *Alice {
	return &Alice{dkg.NewAlice(curve)}
}

type Bob struct {
	*dkg.Bob
}

func NewBob(curve *curves.Curve) *Bob {
	return &Bob{dkg.NewBob(curve)}
}
