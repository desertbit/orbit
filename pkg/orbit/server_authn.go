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

package orbit

import (
	"net"
	"time"

	"github.com/desertbit/orbit/internal/flusher"
)

const (
	flushTimeout = 3 * time.Second
)

// authnSession calls the authentication func defined in the given config.
// If no such func has been defined, it returns immediately.
// This is a convenience func that ensures the connection is flushed,
// even if the authentication fails. That ensures that errors can be
// handled appropriately on each peer's side.
// Can be called for both the server and client side.
func authnSession(conn net.Conn, config *Config) (v interface{}, err error) {
	// Skip if no authentication hook is defined.
	if config.AuthnFunc == nil {
		return
	}

	// Always flush the connection on defer.
	// Otherwise authentication errors might not be send
	// to the peer, because the connection is closed too fast.
	defer func() {
		derr := flusher.Flush(conn, flushTimeout)
		if err == nil {
			err = derr
		}
	}()

	// Reset the deadlines.
	err = conn.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Call the authentication func defined in the config.
	return config.AuthnFunc(conn)
}
