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

package client

import (
	"context"
	"fmt"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/transport"
)

const (
	openCancelStreamTimeout = 10 * time.Second
	cancelRequestTimeout    = 3 * time.Second
)

func (s *session) cancelCall(key uint32) error {
	// Open the cancel stream if not present.
	stream, err := s.openCancelStream()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), cancelRequestTimeout)
	defer cancel()

	// Cancel the call on the remote peer.
	return s.writeRPCRequest(ctx, stream, &s.cancelStreamMx, api.RPCTypeCancel, &api.RPCCall{Key: key}, nil, 0)
}

func (s *session) openCancelStream() (stream transport.Stream, err error) {
	s.cancelStreamMx.Lock()
	defer s.cancelStreamMx.Unlock()

	// Check if the stream is already present.
	if s.cancelStream != nil && !s.cancelStream.IsClosed() {
		stream = s.cancelStream
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), openCancelStreamTimeout)
	defer cancel()

	stream, err = s.openStream(ctx, api.StreamTypeCancelCalls, nil, 0)
	if err != nil {
		err = fmt.Errorf("failed to open cancel stream: %w", err)
		return
	}
	s.cancelStream = stream
	return
}
