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

package packet

import (
	"errors"
	"net"
	"time"

	"github.com/desertbit/orbit/codec"
	"github.com/desertbit/orbit/internal/bytes"
)

var (
	ErrInvalidPayloadSize     = errors.New("invalid payload size")
	ErrMaxPayloadSizeExceeded = errors.New("max payload size exceeded")
)

// ReadDecode reads the packet and decodes it into the value.
// Ensure to pass a pointer value.
func ReadDecode(
	conn net.Conn,
	value interface{},
	codec codec.Codec,
	maxPayloadSize int,
	timeout time.Duration,
) (err error) {
	payload, err := ReadTimeout(conn, nil, maxPayloadSize, timeout)
	if err != nil {
		return
	}

	return codec.Decode(payload, value)
}

// ReadTimeout same as Read, but sets the read deadline.
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

	return Read(conn, buffer, maxPayloadSize)
}

// Read from the connection. Sets the read deadline to the timeout
// and returns the packet bytes. if buffer is set and is big enough to
// fit the packet, then the buffer is used. Otherwise a new buffer
// is allocated.
func Read(
	conn net.Conn,
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
		return nil, ErrInvalidPayloadSize
	} else if payloadLen > maxPayloadSize {
		return nil, ErrMaxPayloadSizeExceeded
	}

	// Create a new buffer if the passed buffer is too small.
	if len(buffer) < payloadLen {
		buffer = make([]byte, payloadLen)
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

// WriteEncode encodes the value and writes it to the connection.
func WriteEncode(
	conn net.Conn,
	value interface{},
	codec codec.Codec,
	maxPayloadSize int,
	timeout time.Duration,
) (err error) {
	data, err := codec.Encode(value)
	if err != nil {
		return
	}

	return WriteTimeout(conn, data, maxPayloadSize, timeout)
}

// Write the packet data to the connection and set the write deadline.
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

	return Write(conn, data, maxPayloadSize)
}

// Write the packet data to the connection.
func Write(
	conn net.Conn,
	data []byte,
	maxPayloadSize int,
) (err error) {
	payloadLen := len(data)
	if payloadLen == 0 {
		return ErrInvalidPayloadSize
	} else if payloadLen > maxPayloadSize {
		return ErrMaxPayloadSizeExceeded
	}

	// Write the payload length.
	payloadBufLen := bytes.FromUint32(uint32(payloadLen))
	_, err = conn.Write(payloadBufLen)
	if err != nil {
		return
	}

	// Write the payload data.
	_, err = conn.Write(data)
	if err != nil {
		return
	}

	return
}
