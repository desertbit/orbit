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
	"github.com/desertbit/orbit/pkg/packet"
	ot "github.com/desertbit/orbit/pkg/transport"
	"github.com/rs/zerolog/log"
)

const (
	maxServiceIDSize = 128
)

// Mux is a wrapper around a transport.Transport and offers custom
// dial and listen methods, which allow to differentiate between
// multiple services, effectively enabling multiplexing multiple services
// over a single transport.
type Mux struct {
	tr ot.Transport

	opts Options

	mx sync.Mutex
	// Client
	dialConn ot.Conn
	// Server
	isListening bool
	listenAddr  net.Addr
	services    map[string]*service
}

// small helper struct representing a registered service.
type service struct {
	connChan   chan ot.Conn   // listener channel.
	streamChan chan ot.Stream // conn channel.
}

// New creates a new multiplexer wrapping the given transport.
// In order to use the Mux as a transport.Transport, call Mux.Transport
// with a suitable serviceID.
func New(tr ot.Transport, opts Options) (*Mux, error) {
	// Set default values, where needed.
	err := options.SetDefaults(&opts, DefaultOptions())
	if err != nil {
		return nil, err
	}

	return &Mux{
		tr:       tr,
		opts:     opts,
		services: make(map[string]*service),
	}, nil
}

// Transport returns a small transport.Transport using m as its multiplexer.
func (m *Mux) Transport(serviceID string) ot.Transport {
	return newTransport(m, serviceID)
}

func (m *Mux) dial(cl closer.Closer, ctx context.Context, serviceID string) (conn ot.Conn, err error) {
	// Establish a basic connection, if not yet done.
	m.mx.Lock()
	if m.dialConn == nil {
		m.dialConn, err = m.tr.Dial(cl, ctx)
		if err != nil {
			m.mx.Unlock()
			return
		}
	}
	m.mx.Unlock()

	// Create a new mux.Conn with it.
	conn = newConn(m.dialConn, serviceID, m.opts.InitTimeout, nil)

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
	m.services[serviceID] = &service{connChan: connChan}

	return ln, nil
}

func (m *Mux) listenRoutine(ln ot.Listener) {
	var (
		conn ot.Conn
		err  error
	)

	ctx, cancel := ln.Context()
	defer cancel()

	// Wait for incoming connections and pass them to our routine.
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
	)

	// Wait for incoming streams on the connection and handle them.
	for {
		stream, err = conn.AcceptStream(ctx)
		if err != nil {
			if !conn.IsClosing() && !conn.IsClosedError(err) {
				log.Error().Err(err).Msg("mux.transport: failed to accept stream")
			}
			return
		}

		go func() {
			err = m.handleStream(conn, stream)
			if err != nil {
				log.Error().Err(err).Msg("mux.transport: failed to handle stream")
			}
		}()
	}
}

// handleStream reads the serviceID off of the given stream, before forwarding it to
// the respective service.
// If its the first stream of the given conn, a new mux conn is properly initialized
// and sent to the service's listener.
func (m *Mux) handleStream(conn ot.Conn, stream ot.Stream) (err error) {
	defer func() {
		if err != nil {
			_ = stream.Close()
		}
	}()

	if m.opts.InitTimeout > 0 {
		err = stream.SetReadDeadline(time.Now().Add(m.opts.InitTimeout))
		if err != nil {
			return fmt.Errorf("failed to set read deadline: %v", err)
		}
	}

	// Read the service id.
	idBuf := make([]byte, maxServiceIDSize)
	data, err := packet.Read(stream, idBuf, maxServiceIDSize)
	if err != nil {
		return
	}
	serviceID := string(data)

	// Reset deadline.
	if m.opts.InitTimeout > 0 {
		err = stream.SetReadDeadline(time.Time{})
		if err != nil {
			return
		}
	}

	var (
		srv *service
		ok  bool
	)

	m.mx.Lock()
	defer m.mx.Unlock()

	// Check, if the service has been registered.
	srv, ok = m.services[serviceID]
	if !ok {
		return fmt.Errorf("service '%s' has not been registered", serviceID)
	} else if srv.streamChan == nil {
		// Create the mux conn and send it to the service once.
		srv.streamChan = make(chan ot.Stream, 3)
		mc := newConn(conn, "", 0, srv.streamChan)

		select {
		case <-conn.ClosingChan():
			return
		case srv.connChan <- mc:
		}
	}

	// Pass the new stream to the service.
	select {
	case <-conn.ClosingChan():
		return
	case srv.streamChan <- stream:
	}

	return
}
