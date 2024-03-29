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

/*
Package packet provides convenience methods to read/write packets to a net.Conn.
A selection of different functions are available to specify timeouts and to provide
either raw bytes or custom values that are en-/decoded with custom codec.Codecs.

A packet is a very simple network data format used to transmit data over
a network. It consists of two parts, a header and the payload.

The header must be present and has a fixed size of 4 bytes. It only contains the size of
the payload.

The payload is of variable size (as indicated by the header) and is a stream of bytes.
The maximum possible size is restricted by the available bytes for the header, and is
therefore currently 2^32 - 1.

In case an empty payload should be transmitted, the packet consists of solely a header
with value 0.
*/
package packet

import (
	"errors"
	"math"
	"net"

	"github.com/desertbit/orbit/internal/bytes"
	"github.com/desertbit/orbit/pkg/codec"
)

const (
	MaxSize            = math.MaxInt32
	NoPayloadSizeLimit = -1
)

var (
	ErrZeroData               = errors.New("zero data")
	ErrMaxPayloadSizeExceeded = errors.New("maximum payload size exceeded")
)

// ReadDecode reads the packet from the connection using Read()
// and decodes it into the value, using the provided codec.
// Ensure to pass a pointer value.
func ReadDecode(conn net.Conn, value interface{}, codec codec.Codec, maxPayloadSize int) error {
	// Read the packet from the connection.
	payload, err := Read(conn, nil, maxPayloadSize)
	if err != nil {
		return err
	}

	// Decode it into the value using the codec.
	return codec.Decode(payload, value)
}

// Read reads a packet from the connection and returns the raw bytes of it.
//
// A maxPayloadSize of 0 specifies no limit.
// A maxPayloadSize of -1 (NoPayloadSizeLimit) specifies no limit.
//
// If buffer is set and is big enough to fit the packet,
// then the buffer is used. Otherwise a new buffer is allocated.
// Returns an empty byte slice when no data was send.
// Returns ErrZeroData, if a packet with length 0 is received.
// Returns ErrMaxPayloadSizeExceeded if the payload size is exceeded.
func Read(conn net.Conn, buffer []byte, maxPayloadSize int) ([]byte, error) {
	var (
		err          error
		n, bytesRead int

		payloadSizeBuf = make([]byte, 4)
	)

	// Read the payload size from the stream.
	for bytesRead < 4 {
		n, err = conn.Read(payloadSizeBuf[bytesRead:])
		if err != nil {
			return nil, err
		}
		bytesRead += n
	}

	// Extract the payload size and verify it.
	payloadLen32, err := bytes.ToUint32(payloadSizeBuf)
	if err != nil {
		return nil, err
	}
	payloadLen := int(payloadLen32)
	if payloadLen == 0 {
		return []byte{}, ErrZeroData
	} else if maxPayloadSize != NoPayloadSizeLimit && payloadLen > maxPayloadSize {
		return nil, ErrMaxPayloadSizeExceeded
	}

	// Create a new buffer if the passed buffer is too small.
	// Otherwise, ensure the length fits.
	if len(buffer) < payloadLen {
		buffer = make([]byte, payloadLen)
	} else if len(buffer) > payloadLen {
		buffer = buffer[:payloadLen]
	}

	// Read the payload from the connection.
	bytesRead = 0
	for bytesRead < payloadLen {
		n, err = conn.Read(buffer[bytesRead:])
		if err != nil {
			return nil, err
		}
		bytesRead += n
	}

	return buffer, nil
}

// WriteEncode encodes the value using the provided codec
// and writes it to the connection using WriteTimeout().
func WriteEncode(conn net.Conn, value interface{}, codec codec.Codec, maxPayloadSize int) error {
	// Encode the value using the codec.
	data, err := codec.Encode(value)
	if err != nil {
		return err
	}

	// Write the data to the connection with a timeout.
	return Write(conn, data, maxPayloadSize)
}

// Write writes the packet data to the connection, without
// setting a timeout.
//
// If data is empty, an empty packet is sent that consists of
// a header with payload size 0 and no payload.
//
// A maxPayloadSize of -1 (NoPayloadSizeLimit) specifies no limit.
//
// Returns ErrMaxPayloadSizeExceeded if the payload size is exceeded.
func Write(conn net.Conn, data []byte, maxPayloadSize int) error {
	payloadLen := len(data)

	if payloadLen > MaxSize || maxPayloadSize != NoPayloadSizeLimit && payloadLen > maxPayloadSize {
		return ErrMaxPayloadSizeExceeded
	}

	// Write the payload length.
	payloadBufLen := bytes.FromUint32(uint32(payloadLen))
	_, err := conn.Write(payloadBufLen)
	if err != nil {
		return err
	}

	// Write the payload data if present.
	if payloadLen > 0 {
		_, err = conn.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}
