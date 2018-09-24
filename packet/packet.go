/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers[at]desertbit.com>
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


/*
Package packet provides convenience methods to read/write packets to a net.Conn.
A selection of different functions are available to specify timeouts and to provide
either raw bytes or custom values that are en-/decoded with custom codec.Codecs.

Packets

A packet is a very simple and efficient network data format used to transmit data over
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
	"io"
	"net"
	"time"

	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/internal/bytes"
)

var (
	// Returned, if the size of a payload exceeds the maxPayloadSize.
	ErrMaxPayloadSizeExceeded = errors.New("max payload size exceeded")
)

// ReadDecode reads the packet from the connection using ReadTimeout()
// and decodes it into the value, using the provided codec.
// Ensure to pass a pointer value.
func ReadDecode(
	conn net.Conn,
	value interface{},
	codec codec.Codec,
	maxPayloadSize int,
	timeout time.Duration,
) (err error) {
	// Read the packet from the connection with a timeout.
	payload, err := ReadTimeout(conn, nil, maxPayloadSize, timeout)
	if err != nil {
		return
	}

	// Decode it into the value using the codec.
	return codec.Decode(payload, value)
}

// ReadTimeout performs the same task as Read(), but allows
// to specify a timeout for reading from the connection.
// If the deadline is not met, the read is terminated and a
// net.Error with Timeout() == true is returned.
func ReadTimeout(
	conn net.Conn,
	buffer []byte,
	maxPayloadSize int,
	timeout time.Duration,
) ([]byte, error) {
	// Set the timeout.
	err := conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, err
	}

	// Read the packet from the connection.
	return Read(conn, buffer, maxPayloadSize)
}

// Read reads a packet from the connection, without setting a timeout,
// and returns the raw bytes of it.
// A maximum size for the packet must be specified. If the packet's size
// exceeds it, an ErrMaxPayloadSizeExceeded is returned.
// If buffer is set and is big enough to fit the packet,
// then the buffer is used. Otherwise a new buffer is allocated.
// Returns an empty byte slice when no data was send.
func Read(
	conn io.Reader,
	buffer []byte,
	maxPayloadSize int,
) ([]byte, error) {
	var (
		err            error
		n, bytesRead   int
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
		return []byte{}, nil
	} else if payloadLen > maxPayloadSize {
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
func WriteEncode(
	conn net.Conn,
	value interface{},
	codec codec.Codec,
	maxPayloadSize int,
	timeout time.Duration,
) (err error) {
	// Encode the value using the codec.
	data, err := codec.Encode(value)
	if err != nil {
		return
	}

	// Write the data to the connection with a timeout.
	return WriteTimeout(conn, data, maxPayloadSize, timeout)
}

// WriteTimeout performs the same task as Write(), but allows
// to specify a timeout for reading from the connection.
// If the deadline is not met, the read is terminated and a
// net.Error with Timeout() == true is returned.
func WriteTimeout(
	conn net.Conn,
	data []byte,
	maxPayloadSize int,
	timeout time.Duration,
) (err error) {
	// Set the write deadline.
	err = conn.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return
	}

	// Write the packet to the connection.
	return Write(conn, data, maxPayloadSize)
}

// Write writes the packet data to the connection, without
// setting a timeout.
// If data is empty, an empty packet is sent that consists of
// a header with payload size 0 and no payload.
// A maximum size for the packet must be specified. If the packet's size
// exceeds it, an ErrMaxPayloadSizeExceeded is returned.
func Write(
	conn io.Writer,
	data []byte,
	maxPayloadSize int,
) (err error) {
	payloadLen := len(data)
	if payloadLen > maxPayloadSize {
		return ErrMaxPayloadSizeExceeded
	}

	// Write the payload length.
	payloadBufLen := bytes.FromUint32(uint32(payloadLen))
	_, err = conn.Write(payloadBufLen)
	if err != nil {
		return
	}

	// Write the payload data if present.
	if payloadLen > 0 {
		_, err = conn.Write(data)
		if err != nil {
			return
		}
	}

	return
}
