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

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/internal/flusher"
	"github.com/desertbit/orbit/internal/utils"
	"github.com/desertbit/orbit/pkg/packet"
)

const (
	// The length of the randomly created session ids.
	sessionIDLength = 32
)

func newServerSession(conn Conn, cf *Config, h SessionHandler, hs []Hook) (sn *Session, err error) {
	// Close connection on error.
	defer func() {
		if err != nil {
			conn.Close_()
		}
	}()

	// Generate an id for the session.
	id, err := utils.RandomString(sessionIDLength)
	if err != nil {
		return nil, err
	}

	// Create new context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), cf.InitTimeout)
	defer cancel()

	// Wait for an incoming stream from the client.
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		fErr := flusher.Flush(stream, flushTimeout)
		if err == nil {
			err = fErr
		}
		_ = stream.Close()
	}()

	// Deadline is certainly available.
	deadline, _ := ctx.Deadline()

	// Set the deadline for the handshake.
	err = stream.SetDeadline(deadline)
	if err != nil {
		return nil, err
	}

	// Wait for the client's handshake request.
	var args api.HandshakeArgs
	err = packet.ReadDecode(stream, &args, cf.Codec)
	if err != nil {
		return nil, err
	}

	// Set the response data for the handshake.
	var ret api.HandshakeRet
	if args.Version != api.Version {
		ret.Code = api.HSInvalidVersion
	} else {
		ret.Code = api.HSOk
		ret.SessionID = id
	}

	// Send the handshake response back to the client.
	err = packet.WriteEncode(stream, &ret, cf.Codec)
	if err != nil {
		return nil, err
	}

	// Abort, if the handshake failed.
	if ret.Code == api.HSInvalidVersion {
		return nil, ErrInvalidVersion
	} else if ret.Code != api.HSOk {
		return nil, fmt.Errorf("unknown handshake code %d", ret.Code)
	}

	// Reset the deadline.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return nil, err
	}

	// Finally, create the orbit session.
	// Hooks will be called, which are the last chance that the
	// session might not be established.
	return newSession(conn, stream, id, cf, h, hs)
}
