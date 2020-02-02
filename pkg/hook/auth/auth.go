/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2020 Sebastian Borchers <sebastian[at]desertbit.com>
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
	"errors"
	"time"

	"github.com/desertbit/orbit/pkg/hook/auth/bcrypt"
)

const (
	timeout      = 20 * time.Second
	flushTimeout = 7 * time.Second
)

var (
	errAuthFailed      = errors.New("authentication failed")
	errInvalidUsername = errors.New("invalid username")
	errInvalidPassword = errors.New("invalid password")
)

// Checksum calculates the SHA256 Checksum of the given string.
func Checksum(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}

// Hash calculates the bcrypt hash of the SHA256 checksum of the given string.
func Hash(s string) ([]byte, error) {
	checksum := Checksum(s)
	return bcrypt.Generate(checksum[:])
}
