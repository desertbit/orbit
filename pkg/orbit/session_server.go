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
	"context"
	"errors"
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/internal/api"
)

// serverSession is the internal helper to initialize a new server-side session.
func newServerSession(cl closer.Closer, conn Conn, cf *Config) (s *Session, err error) {
	// Always close the conn on error.
	defer func() {
		if err != nil {
			conn.Close_()
		}
	}()

	// Prepare the config with default values, where needed.
	cf = prepareConfig(cf)

	// Create new context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), cf.InitTimeout)
	defer cancel()

	// Wait for an incoming stream.
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return
	}

	// Deadline is certainly available.
	deadline, _ := ctx.Deadline()

	// Set a read timeout.
	err = stream.SetReadDeadline(deadline)
	if err != nil {
		return
	}

	// Ensure client has the same protocol version.
	v := make([]byte, 1)
	n, err := stream.Read(v)
	if err != nil {
		return
	} else if n != 1 {
		return nil, errors.New("failed to read version byte from connection")
	} else if v[0] != api.Version {
		return nil, ErrInvalidVersion
	}

	// Reset the deadline.
	err = stream.SetReadDeadline(time.Time{})
	if err != nil {
		return
	}

	// TODO: auth hook?

	// Finally, create the orbit server session.
	s = newSession(cl, conn, stream, cf)

	// TODO: remove?
	// Save the arbitrary data from the auth func.
	//s.Value = value TODO
	return
}
