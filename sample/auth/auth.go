/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package auth

import (
	"crypto/sha256"
	"golang.org/x/crypto/bcrypt"
	"time"

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
