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
	"fmt"
	"time"

	"github.com/desertbit/orbit/pkg/packet"
	ot "github.com/desertbit/orbit/pkg/transport"
)

type conn struct {
	ot.Conn

	// Client
	serviceID   []byte
	initTimeout time.Duration
	// Server
	streamChan <-chan ot.Stream
}

func newClientConn(tc ot.Conn, serviceID string, initTimeout time.Duration) *conn {
	return &conn{
		Conn:        tc,
		serviceID:   []byte(serviceID),
		initTimeout: initTimeout,
	}
}

func newServerConn(tc ot.Conn, streamChan <-chan ot.Stream) *conn {
	return &conn{
		Conn:       tc,
		streamChan: streamChan,
	}
}

// Implements the transport.Conn interface.
func (c *conn) OpenStream(ctx context.Context) (stream ot.Stream, err error) {
	stream, err = c.Conn.OpenStream(ctx)
	if err != nil {
		err = fmt.Errorf("failed to open stream: %v", err)
		return
	}

	if c.initTimeout > 0 {
		err = stream.SetWriteDeadline(time.Now().Add(c.initTimeout))
		if err != nil {
			err = fmt.Errorf("failed to set read deadline: %v", err)
			return
		}
	}

	// Send the service id first.
	err = packet.Write(stream, c.serviceID, maxServiceIDSize)
	if err != nil {
		err = fmt.Errorf("failed to send service id: %v", err)
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
