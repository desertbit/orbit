/*
 * Guru - nLab
 * Copyright (c) 2020 Wahtari GmbH
 */

//go:generate msgp
package auth

import (
	"crypto"
	"crypto/rand"
	"errors"
	"time"

	"github.com/desertbit/orbit/pkg/codec"
)

// Sign signs the given Token with the given signer and its options,
// producing a byte slice containing both token and signature.
func Sign(t *Token, cs crypto.Signer, csOpts crypto.SignerOpts, cc codec.Codec) (data []byte, err error) {
	// Encode the token.
	data, err = cc.Encode(t)
	if err != nil {
		return
	}

	// Create the signature of the token.
	sig, err := cs.Sign(rand.Reader, data, csOpts)
	if err != nil {
		return nil, err
	}

	// Append the signature to the data.
	data = append(data, sig...)
	return
}

// Tags are as short as possible to reduce size on wire.
type Token struct {
	UserID    string      `msgpack:"u" json:"u"`
	IssuedOn  time.Time   `msgpack:"i" json:"i"`
	ExpiresOn time.Time   `msgpack:"e" json:"e"`
	Data      interface{} `msgpack:"d" json:"d"`
}

func (t *Token) Valid(ti time.Time) error {
	if t.UserID == "" {
		return errors.New("empty user id")
	}
	if t.IssuedOn.IsZero() {
		return errors.New("invalid issued on claim")
	}
	if t.IssuedOn.After(ti) {
		return errors.New("issued on claim is in the future")
	}
	if t.ExpiresOn.IsZero() {
		return errors.New("invalid expires on claim")
	}
	if t.ExpiresOn.Before(ti) {
		return errors.New("expired")
	}
	return nil
}
