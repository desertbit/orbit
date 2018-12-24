/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2018 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2018 Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package auth

import (
	"crypto/sha256"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/pkg/errors"
)

const (
	readTimeout    = 12 * time.Second
	writeTimeout   = 12 * time.Second
	maxPayloadSize = 2 * 1024 // 2KB
)

// WARNING! JUST A SHOWCASE ---------------------------------
const (
	SampleUsername = "sampleUser"
	SamplePassword = "samplePassword"
)

var SamplePasswordHash = []byte("$2a$10$A3090fBqvJ3SPg0VGf2afePobWmEAQZ0VbTzbnVLNVCe5blymBham")

// ----------------------------------------------------------

var (
	errAuthFailed      = errors.New("authentication failed")
	errInvalidUsername = errors.New("invalid username")
	errInvalidPassword = errors.New("invalid password")
)

type HashedCredentials map[string][]byte

func PasswordHash(pw string) ([]byte, error) {
	checksum := Checksum(pw)
	return bcrypt.GenerateFromPassword(checksum[:], bcrypt.DefaultCost)
}

func Checksum(pw string) [32]byte {
	return sha256.Sum256([]byte(pw))
}
