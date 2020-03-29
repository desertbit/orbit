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

package rpc

import (
	"errors"
	"fmt"
	"net"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/packet"
)

// Read the rpc request from the connection,
// If a payload or header buffer is passed, then this buffer will be used
// if the capacity is sufficient.
func Read(
	conn net.Conn,
	headerBuffer []byte,
	payloadBuffer []byte,
	maxHeaderSize int,
	maxPayloadSize int,
) (
	reqType api.RPCType,
	header []byte,
	payload []byte,
	err error,
) {
	var (
		bytesRead, n int
		reqTypeBuf   = make([]byte, 1)
	)

	// Read the reqType from the stream.
	for bytesRead == 0 {
		n, err = conn.Read(reqTypeBuf)
		if err != nil {
			return
		}
		bytesRead += n
	}

	reqType = api.RPCType(reqTypeBuf[0])

	// Read the header from the stream.
	header, err = packet.Read(conn, headerBuffer, maxHeaderSize)
	if err != nil {
		err = fmt.Errorf("failed to read header: %w", err)
		return
	}

	// Read the optional payload from the stream.
	payload, err = packet.Read(conn, payloadBuffer, maxPayloadSize)
	if err != nil {
		if errors.Is(err, packet.ErrZeroData) {
			err = nil
		} else {
			err = fmt.Errorf("failed to read payload: %w", err)
			return
		}
	}

	return
}

// Write the rpc request to the connection.
func Write(
	conn net.Conn,
	reqType api.RPCType,
	header []byte,
	payload []byte,
	maxHeaderSize int,
	maxPayloadSize int,
) (err error) {
	// Write the request type.
	_, err = conn.Write([]byte{byte(reqType)})
	if err != nil {
		return
	}

	// Write the header.
	err = packet.Write(conn, header, maxHeaderSize)
	if err != nil {
		err = fmt.Errorf("failed to write header: %w", err)
		return
	}

	// Write the payload.
	err = packet.Write(conn, payload, maxPayloadSize)
	if err != nil {
		err = fmt.Errorf("failed to write payload: %w", err)
		return
	}
	return
}
