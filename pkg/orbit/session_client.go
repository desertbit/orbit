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
	"fmt"
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/packet"
)

// newClientSession is the internal helper to initialize a new client-side session.
func newClientSession(cl closer.Closer, conn Conn, cf *Config, hs []Hook) (s *Session, err error) {
	// Always close the conn on error.
	defer func() {
		if err != nil {
			conn.Close_()
		}
	}()

	// Prepare the config with default values, where needed.
	cf = prepareConfig(cf)

	// Create new timeout context.
	ctx, cancel := context.WithTimeout(context.Background(), cf.InitTimeout)
	defer cancel()

	// Open the stream to the server.
	stream, err := conn.OpenStream(ctx)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	// Deadline is certainly available.
	deadline, _ := ctx.Deadline()

	// Set the deadline for the handshake.
	err = stream.SetDeadline(deadline)
	if err != nil {
		return nil, err
	}

	// Send the arguments for the handshake to the server.
	err = packet.WriteEncode(stream, &api.HandshakeArgs{Version: api.Version}, cf.Codec)
	if err != nil {
		return nil, err
	}

	// Wait for the server's handshake response.
	var ret api.HandshakeRet
	err = packet.ReadDecode(stream, &ret, cf.Codec)
	if err != nil {
		return nil, err
	} else if ret.Code == api.HSInvalidVersion {
		return nil, ErrInvalidVersion
	} else if ret.Code != api.HSOk {
		return nil, fmt.Errorf("unknown handshake code %d", ret.Code)
	}

	// Reset the deadline.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	// Finally, create the orbit session.
	// Hooks will be called, which are the last chance that the
	// session might not be established.
	return newSession(cl, conn, stream, ret.SessionID, cf, hs)
}
