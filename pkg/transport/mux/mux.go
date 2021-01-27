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

type Mux struct {
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

func New(tr ot.Transport, opts Options) (*Mux, error) {
	// Set default values, where needed.
	err := options.SetDefaults(&opts, DefaultOptions())
	if err != nil {
		return nil, err
	}

	return &Mux{
		tr:       tr,
		opts:     opts,
		services: make(map[string]chan ot.Conn),
		streams:  make(map[uint16]chan ot.Stream),
	}, nil
}

func (m *Mux) Transport(serviceID string) ot.Transport {
	return newTransport(m, serviceID)
}

func (m *Mux) dial(cl closer.Closer, ctx context.Context, serviceID string) (conn ot.Conn, err error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	// Establish a basic connection, if not yet done.
	if m.dialConn == nil {
		m.dialConn, err = m.tr.Dial(cl, ctx)
		if err != nil {
			return
		}
	}

	// Create a new mux.Conn with it.
	mc := newConn(m.dialConn, m.dialStreamID, nil)
	m.dialStreamID++

	// Initialize this service.
	err = mc.initStream(ctx, serviceID, m.opts.InitTimeout)
	if err != nil {
		return
	}

	conn = mc
	return
}

func (m *Mux) listen(cl closer.Closer, serviceID string) (ot.Listener, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if !m.isListening {
		// Create our listener.
		ln, err := m.tr.Listen(cl)
		if err != nil {
			return nil, err
		}

		go m.listenRoutine(ln)

		m.isListening = true
		m.listenAddr = ln.Addr()
	}

	// Check, if the service has been registered already.
	if _, ok := m.services[serviceID]; ok {
		return nil, fmt.Errorf("serviceID '%s' registered twice", serviceID)
	}

	// Create a new service listener.
	connChan := make(chan ot.Conn, 3)
	ln := newListener(cl, connChan, m.listenAddr)
	m.services[serviceID] = connChan

	return ln, nil
}

func (m *Mux) listenRoutine(ln ot.Listener) {
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

		go m.acceptStreamRoutine(ctx, conn)
	}
}

func (m *Mux) acceptStreamRoutine(ctx context.Context, conn ot.Conn) {
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

		err = m.handleStream(conn, stream, typeBuf, idBuf)
		if err != nil {
			log.Error().Err(err).Msg("mux.transport: failed to handle stream")
			return
		}
	}
}

func (m *Mux) handleStream(conn ot.Conn, stream ot.Stream, typeBuf, idBuf []byte) (err error) {
	if m.opts.InitTimeout > 0 {
		err = stream.SetReadDeadline(time.Now().Add(m.opts.InitTimeout))
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
		m.mx.Lock()
		connChan, ok = m.services[header.ServiceID]
		m.mx.Unlock()
		if !ok {
			return fmt.Errorf("service '%s' not registered", header.ServiceID)
		}

		// Create a new mux conn + a stream channel for it.
		streamChan := make(chan ot.Stream, 3)
		mc := newConn(conn, header.StreamID, streamChan)
		m.mx.Lock()
		m.streams[header.StreamID] = streamChan
		m.mx.Unlock()

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
		m.mx.Lock()
		streamChan, ok = m.streams[streamID]
		m.mx.Unlock()
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
