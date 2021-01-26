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
	"errors"
	"net"

	"github.com/desertbit/closer/v3"
	ot "github.com/desertbit/orbit/pkg/transport"
)

type listener struct {
	closer.Closer

	connChan chan ot.Conn

	addr net.Addr
}

func newListener(cl closer.Closer, connChan chan ot.Conn, addr net.Addr) ot.Listener {
	return &listener{
		Closer:   cl,
		connChan: connChan,
		addr:     addr,
	}
}

func (ln *listener) Accept() (conn ot.Conn, err error) {
	select {
	case <-ln.ClosingChan():
		err = errors.New("closed")
	case conn = <-ln.connChan:
	}
	return
}

// Addr returns the listener's network address.
func (ln *listener) Addr() net.Addr {
	return ln.addr
}
