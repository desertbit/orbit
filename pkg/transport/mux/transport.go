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

	mx          sync.Mutex
	isListening bool
	listenAddr  net.Addr
	dialConn    ot.Conn
	services    map[string]chan ot.Conn
}

func NewTransport(tr ot.Transport, opts Options) (t ot.Transport, err error) {
	// Set default values, where needed.
	err = options.SetDefaults(&opts, DefaultOptions("", ""))
	if err != nil {
		return
	}

	err = opts.validate()
	if err != nil {
		return
	}

	t = &transport{
		tr:       tr,
		opts:     opts,
		services: make(map[string]chan ot.Conn, 5),
	}
	return
}

// Implements the transport.Transport interface.
func (t *transport) Dial(cl closer.Closer, ctx context.Context, v interface{}) (conn ot.Conn, err error) {
	vv, ok := v.(*value)
	if !ok {
		return nil, errors.New("could not cast v to *value; use Value() to properly pass a correct value to Dial()")
	}

	t.mx.Lock()
	defer t.mx.Unlock()

	if t.dialConn == nil {
		t.dialConn, err = t.tr.Dial(cl, ctx, t.opts.DialAddr)
		if err != nil {
			return
		}
	}

	// Send the service identifier.
	stream, err := t.dialConn.OpenStream(ctx)
	if err != nil {
		return
	}

	if t.opts.WriteTimeout > 0 {
		err = stream.SetWriteDeadline(time.Now().Add(t.opts.WriteTimeout))
		if err != nil {
			return
		}
	}

	err = packet.Write(stream, []byte(vv.serviceID), maxPayloadSize)
	if err != nil {
		return
	}

	conn = t.dialConn
	return
}

// Implements the transport.Transport interface.
func (t *transport) Listen(cl closer.Closer, v interface{}) (ot.Listener, error) {
	vv, ok := v.(*value)
	if !ok {
		return nil, errors.New("could not cast v to *value; use Value() to properly pass a correct value to Dial()")
	}

	t.mx.Lock()
	defer t.mx.Unlock()

	if !t.isListening {
		// Create our listener.
		ln, err := t.tr.Listen(cl, t.opts.ListenAddr)
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

	// Create  a new service listener.
	connChan := make(chan ot.Conn, 3)
	ln := newListener(cl, connChan, t.listenAddr)
	t.services[vv.serviceID] = connChan

	return ln, nil
}

func (t *transport) listenRoutine(ln ot.Listener) {
	var (
		conn ot.Conn
		err  error

		buf         = make([]byte, 128)
		closingChan = ln.ClosingChan()
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

		err = t.listenHandler(ctx, closingChan, conn, buf)
		if err != nil {
			if !ln.IsClosing() {
				log.Error().Err(err).Msg("listenRoutine: listenHandler")
				continue
			}
			return
		}
	}
}

func (t *transport) listenHandler(ctx context.Context, closingChan <-chan struct{}, conn ot.Conn, buf []byte) (err error) {
	// Read the service identifier off of the connection.
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to accept stream: %v", err)
	}

	if t.opts.ReadTimeout > 0 {
		err = stream.SetReadDeadline(time.Now().Add(t.opts.ReadTimeout))
		if err != nil {
			return fmt.Errorf("failed to set read deadline: %v", err)
		}
	}

	data, err := packet.Read(stream, buf, maxPayloadSize)
	if err != nil {
		return fmt.Errorf("failed to read packet: %v", err)
	}

	// Service identifier.
	serviceID := string(data)

	var (
		connChan chan ot.Conn
		ok       bool
	)
	t.mx.Lock()
	connChan, ok = t.services[serviceID]
	t.mx.Unlock()

	if !ok {
		// Close the connection.
		conn.Close_()
		log.Warn().Str("serviceID", serviceID).Msg("service not registered, but tried to connect")
	}

	// Pass the new connection to the service listener.
	select {
	case <-closingChan:
	case connChan <- conn:
	}

	return
}
