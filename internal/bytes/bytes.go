/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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

package bytes

import (
	"encoding/binary"
	"errors"
)

//#################//
//### Variables ###//
//#################//

var (
	// Endian defines the byte order used for the encoding.
	Endian binary.ByteOrder = binary.BigEndian

	ErrInvalidLen = errors.New("invalid byte length")
)

//##############//
//### Public ###//
//##############//

// ToUint16 encodes the byte slice to an uint16.
func ToUint16(data []byte) (v uint16, err error) {
	if len(data) < 2 {
		return 0, ErrInvalidLen
	}
	v = Endian.Uint16(data)
	return
}

// FromUint16 converts an uint16 to a byte slice.
func FromUint16(v uint16) (data []byte) {
	data = make([]byte, 2)
	Endian.PutUint16(data, v)
	return
}

// ToUint32 encodes the byte slice to an uint32.
func ToUint32(data []byte) (v uint32, err error) {
	if len(data) < 4 {
		return 0, ErrInvalidLen
	}
	v = Endian.Uint32(data)
	return
}

// FromUint32 converts an uint32 to a byte slice.
func FromUint32(v uint32) (data []byte) {
	data = make([]byte, 4)
	Endian.PutUint32(data, v)
	return
}

// ToUint64 encodes the byte slice to an uint64.
func ToUint64(data []byte) (v uint64, err error) {
	if len(data) < 8 {
		return 0, ErrInvalidLen
	}
	v = Endian.Uint64(data)
	return
}

// FromUint64 converts an uint64 to a byte slice.
func FromUint64(v uint64) (data []byte) {
	data = make([]byte, 8)
	Endian.PutUint64(data, v)
	return
}
