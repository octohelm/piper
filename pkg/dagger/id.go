package dagger

import (
	"crypto/rand"
	"io"
	"math/big"
)

var idReader = rand.Reader

const (
	randomIDEntropyBytes = 17
	randomIDBase         = 36
	maxRandomIDLength    = 25
)

func NewID() string {
	var p [randomIDEntropyBytes]byte

	if _, err := io.ReadFull(idReader, p[:]); err != nil {
		panic("failed to read random bytes: " + err.Error())
	}

	p[0] |= 0x80 // set high bit to avoid the need for padding
	return (&big.Int{}).SetBytes(p[:]).Text(randomIDBase)[1 : maxRandomIDLength+1]
}
