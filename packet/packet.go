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

const (
	defaultStreamBufferSize = 32 * 1024 // 32KB (also used in io.Copy())
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
//
// A maximum size for the packet must be specified. If the packet's size
// exceeds it, an ErrMaxPayloadSizeExceeded is returned.
// A negative maxPayloadSize causes a panic.
//
// If buffer is set and is big enough to fit the packet,
// then the buffer is used. Otherwise a new buffer is allocated.
// Returns an empty byte slice when no data was send.
func Read(
	conn io.Reader,
	buffer []byte,
	maxPayloadSize int,
) ([]byte, error) {
	// Check sanity of maxPayloadSize.
	if maxPayloadSize < 0 {
		panic("packet read: maxPayloadSize must be greater or equal than 0")
	}

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
//
// If data is empty, an empty packet is sent that consists of
// a header with payload size 0 and no payload.
//
// A maximum size for the packet must be specified. If the packet's size
// exceeds it, an ErrMaxPayloadSizeExceeded is returned.
// A negative maxPayloadSize causes a panic.
func Write(
	conn io.Writer,
	data []byte,
	maxPayloadSize int,
) (err error) {
	// Check sanity of maxPayloadSize.
	if maxPayloadSize < 0 {
		panic("packet write: maxPayloadSize must be greater or equal than 0")
	}

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

//##############//
//### Stream ###//
//##############//

// ReadStream is a convenience func for ReadStreamBufTimeout.
func ReadStream(
	conn net.Conn,
	dst io.Writer,
) (err error) {
	return ReadStreamBufTimeout(conn, dst, nil, 0)
}

// If buf is nil, a new buffer with default size is allocated.
// A zero or negative timeout means no timeout is set.
// Returns io.ErrShortWrite, if not all data read from
// the connection could be written to dst.
func ReadStreamBufTimeout(
	conn net.Conn,
	dst io.Writer,
	buf []byte,
	timeout time.Duration,
) error {
	var (
		n    int
		data []byte
		err  error

		size = len(buf)
	)

	// If no buffer, or an empty one, has been given, allocate one.
	if size == 0 {
		size = defaultStreamBufferSize
		buf = make([]byte, size)
	}

	for {
		// Read the next packet off the connection.
		// Use a timeout for the read operation, if wanted.
		if timeout > 0 {
			data, err = ReadTimeout(conn, buf, size, timeout)
		} else {
			data, err = Read(conn, buf, size)
		}
		if err != nil {
			return err
		}

		// If an empty packet has been sent, the stream is over.
		if len(data) == 0 {
			return nil
		}

		// Write the data to the destination.
		n, err = dst.Write(data)
		if err != nil {
			return err
		} else if n != len(data) {
			return io.ErrShortWrite
		}
	}
}

// WriteStream is a convenience func for WriteStreamBufTimeout.
func WriteStream(
	conn net.Conn,
	src io.Reader,
) (err error) {
	return WriteStreamBufTimeout(conn, src, nil, 0)
}

// If buf is nil, a new buffer with default size is allocated.
// A zero or negative timeout means no timeout is set.
func WriteStreamBufTimeout(
	conn net.Conn,
	src io.Reader,
	buf []byte,
	timeout time.Duration,
) error {
	var (
		n         int
		err, rErr error

		size = len(buf)
	)

	// If no buffer, or an empty one, has been given, allocate one.
	if size == 0 {
		size = defaultStreamBufferSize

		// If the source is a limited reader, make sure we stay within its limits.
		if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf = make([]byte, size)
	}

	for {
		// Read packet size many bytes from the src.
		n, rErr = src.Read(buf)

		// Process any read bytes before handling the error.
		if n > 0 {
			// Write the data to the connection.
			// Use a timeout on the write operation, if wanted.
			if timeout > 0 {
				err = WriteTimeout(conn, buf[:n], size, timeout)
			} else {
				err = Write(conn, buf[:n], size)
			}
			if err != nil {
				return err
			}
		}

		// Handle the potential read error.
		if rErr == io.EOF {
			// All stream data has been written,
			// send the empty packet to tell its over.
			if timeout > 0 {
				err = WriteTimeout(conn, nil, size, timeout)
			} else {
				err = Write(conn, nil, size)
			}
			if err != nil {
				return err
			}
			return nil
		} else if rErr != nil {
			return rErr
		}
	}
}
