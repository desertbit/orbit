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

/*
Package bytes offers convenience functions to convert bytes
to and from unsigned integers, respecting a defined byte-order.

The underlying byte order and conversion methods being used stem
from the encoding/binary package.
*/
package bytes

import (
	"encoding/binary"
	"errors"
)

var (
	// Endian defines the byte order used for the encoding.
	Endian binary.ByteOrder = binary.BigEndian
	// ErrInvalidLen is an error indicating that a byte slice had invalid length
	// for the conversion that should have been performed.
	ErrInvalidLen = errors.New("invalid byte length")
)

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
