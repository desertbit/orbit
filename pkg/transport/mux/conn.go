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

package mux

import (
	"context"
	"errors"
	"time"

	"github.com/desertbit/orbit/internal/bytes"
	"github.com/desertbit/orbit/pkg/codec/msgpack"
	"github.com/desertbit/orbit/pkg/packet"
	ot "github.com/desertbit/orbit/pkg/transport"
)

const (
	maxInitHeaderSize = 256
)

const (
	initStream byte = 0
	openStream byte = 1
)

type conn struct {
	ot.Conn

	streamID   uint16
	streamChan <-chan ot.Stream
}

func newConn(tc ot.Conn, streamID uint16, streamChan <-chan ot.Stream) *conn {
	return &conn{
		Conn:       tc,
		streamID:   streamID,
		streamChan: streamChan,
	}
}

func (c *conn) initStream(ctx context.Context, serviceID string, initTimeout time.Duration) (err error) {
	stream, err := c.Conn.OpenStream(ctx)
	if err != nil {
		return
	}

	if initTimeout > 0 {
		err = stream.SetWriteDeadline(time.Now().Add(initTimeout))
		if err != nil {
			return
		}
	}

	// Send the stream type first.
	err = packet.Write(stream, []byte{initStream}, 1)
	if err != nil {
		return
	}

	// Register the new service.
	header := initStreamHeader{
		ServiceID: serviceID,
		StreamID:  c.streamID,
	}
	return packet.WriteEncode(stream, header, msgpack.Codec, maxInitHeaderSize)
}

// Implements the transport.Conn interface.
func (c *conn) OpenStream(ctx context.Context) (stream ot.Stream, err error) {
	stream, err = c.Conn.OpenStream(ctx)
	if err != nil {
		return
	}

	// Send the stream type first.
	err = packet.Write(stream, []byte{openStream}, 1)
	if err != nil {
		return
	}

	// Send the stream id to identify the service.
	err = packet.Write(stream, bytes.FromUint16(c.streamID), 2)
	if err != nil {
		return
	}

	return
}

// Implements the transport.Conn interface.
func (c *conn) AcceptStream(ctx context.Context) (stream ot.Stream, err error) {
	select {
	case <-c.ClosingChan():
		err = errors.New("closed")
	case stream = <-c.streamChan:
	}
	return
}
