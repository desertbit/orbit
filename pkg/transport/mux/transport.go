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
	"net"
	"sync"
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/options"
	"github.com/desertbit/orbit/internal/bytes"
	"github.com/desertbit/orbit/pkg/codec/msgpack"
	"github.com/desertbit/orbit/pkg/packet"
	ot "github.com/desertbit/orbit/pkg/transport"
	"github.com/rs/zerolog/log"
)

const (
	maxPayloadSize = 128
)

type transport struct {
	tr ot.Transport

	opts Options

	mx sync.Mutex
	// Client
	dialStreamID uint16
	dialConn     ot.Conn
	// Server
	isListening bool
	listenAddr  net.Addr
	services    map[string]chan ot.Conn
	streams     map[uint16]chan ot.Stream
}

func NewTransport(tr ot.Transport, opts Options) (t ot.Transport, err error) {
	// Set default values, where needed.
	err = options.SetDefaults(&opts, DefaultOptions())
	if err != nil {
		return
	}

	t = &transport{
		tr:       tr,
		opts:     opts,
		services: make(map[string]chan ot.Conn),
		streams:  make(map[uint16]chan ot.Stream),
	}
	return
}

// Implements the transport.Transport interface.
func (t *transport) Dial(cl closer.Closer, ctx context.Context, v interface{}) (conn ot.Conn, err error) {
	vv, ok := v.(*value)
	if !ok {
		return nil, errors.New("could not cast v to mux value; use Value() to properly pass a correct value to Dial()")
	}

	t.mx.Lock()
	defer t.mx.Unlock()

	// Establish a basic connection, if not yet done.
	if t.dialConn == nil {
		t.dialConn, err = t.tr.Dial(cl, ctx, vv.subValue)
		if err != nil {
			return
		}
	}

	// Create a new mux.Conn with it.
	mc := newConn(t.dialConn, t.dialStreamID, nil)
	t.dialStreamID++

	// Initialize this service.
	err = mc.initStream(ctx, vv.serviceID, t.opts.InitTimeout)
	if err != nil {
		return
	}

	conn = mc
	return
}

// Implements the transport.Transport interface.
func (t *transport) Listen(cl closer.Closer, v interface{}) (ot.Listener, error) {
	vv, ok := v.(*value)
	if !ok {
		return nil, errors.New("could not cast v to mux value; use Value() to properly pass a correct value to Dial()")
	}

	t.mx.Lock()
	defer t.mx.Unlock()

	if !t.isListening {
		// Create our listener.
		ln, err := t.tr.Listen(cl, vv.subValue)
		if err != nil {
			return nil, err
		}

		go t.listenRoutine(ln)

		t.isListening = true
		t.listenAddr = ln.Addr()
	}

	// Check, if the service has been registered already.
	if _, ok := t.services[vv.serviceID]; ok {
		return nil, fmt.Errorf("serviceID '%s' registered twice", vv.serviceID)
	}

	// Create a new service listener.
	connChan := make(chan ot.Conn, 3)
	ln := newListener(cl, connChan, t.listenAddr)
	t.services[vv.serviceID] = connChan

	return ln, nil
}

func (t *transport) listenRoutine(ln ot.Listener) {
	var (
		conn ot.Conn
		err  error
	)

	ctx, cancel := ln.Context()
	defer cancel()

	for {
		conn, err = ln.Accept()
		if err != nil {
			if !ln.IsClosing() {
				log.Error().Err(err).Msg("listenRoutine: accept")
			}
			return
		}

		go t.acceptStreamRoutine(ctx, conn)
	}
}

func (t *transport) acceptStreamRoutine(ctx context.Context, conn ot.Conn) {
	defer conn.Close_()

	var (
		stream ot.Stream
		err    error

		typeBuf = make([]byte, 1)
		idBuf   = make([]byte, 2)
	)

	for {
		stream, err = conn.AcceptStream(ctx)
		if err != nil {
			if !conn.IsClosing() && !conn.IsClosedError(err) {
				log.Error().Err(err).Msg("mux.transport: failed to accept stream")
			}
			return
		}

		err = t.handleStream(conn, stream, typeBuf, idBuf)
		if err != nil {
			log.Error().Err(err).Msg("mux.transport: failed to handle stream")
			return
		}
	}
}

func (t *transport) handleStream(conn ot.Conn, stream ot.Stream, typeBuf, idBuf []byte) (err error) {
	if t.opts.InitTimeout > 0 {
		err = stream.SetReadDeadline(time.Now().Add(t.opts.InitTimeout))
		if err != nil {
			return fmt.Errorf("failed to set read deadline: %v", err)
		}
	}

	// Read the stream type.
	data, err := packet.Read(stream, typeBuf, len(typeBuf))
	if err != nil {
		return
	}

	switch data[0] {
	case initStream:
		// Read the stream init header.
		var header initStreamHeader
		err = packet.ReadDecode(stream, &header, msgpack.Codec, maxInitHeaderSize)
		if err != nil {
			return fmt.Errorf("failed to read init stream header: %v", err)
		}

		// Check, if the service is listening.
		var (
			connChan chan ot.Conn
			ok       bool
		)
		t.mx.Lock()
		connChan, ok = t.services[header.ServiceID]
		t.mx.Unlock()
		if !ok {
			return fmt.Errorf("service '%s' not registered", header.ServiceID)
		}

		// Create a new mux conn + a stream channel for it.
		streamChan := make(chan ot.Stream, 3)
		mc := newConn(conn, header.StreamID, streamChan)
		t.mx.Lock()
		t.streams[header.StreamID] = streamChan
		t.mx.Unlock()

		// Pass the mux conn to the listener that has been registered.
		select {
		case <-conn.ClosingChan():
			return
		case connChan <- mc:
		}

	case openStream:
		// Read the stream id.
		data, err = packet.Read(stream, idBuf, len(idBuf))
		if err != nil {
			return
		}

		// Convert to uint16.
		var streamID uint16
		streamID, err = bytes.ToUint16(data)
		if err != nil {
			return
		}

		// Retrieve the stream channel of the associated mux conn.
		var (
			streamChan chan ot.Stream
			ok         bool
		)
		t.mx.Lock()
		streamChan, ok = t.streams[streamID]
		t.mx.Unlock()
		if !ok {
			return fmt.Errorf("unknown stream id %d", streamID)
		}

		// Pass the new stream to this connection.
		select {
		case <-conn.ClosingChan():
			return
		case streamChan <- stream:
		}

	default:
		return fmt.Errorf("unknown stream type %v", data[0])
	}

	return
}
