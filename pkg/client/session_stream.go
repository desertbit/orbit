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
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/packet"
	"github.com/desertbit/orbit/pkg/transport"
)

func (s *session) OpenTypedStream(ctx context.Context, id string, maxArgSize, maxRetSize int) (ts TypedRWStream, err error) {
	// Create a new client context.
	cctx := newContext(ctx, s)

	// Call the OnStream hooks.
	err = s.handler.hookOnStream(cctx, id)
	if err != nil {
		return
	}

	// Call the closed hook if an error occurs.
	defer func() {
		if err != nil {
			s.handler.hookOnStreamClosed(cctx, id, err)
		}
	}()

	// Open the stream.
	stream, err := s.openStream(ctx, api.StreamTypeRaw, &api.StreamRaw{
		ID:   id,
		Data: cctx.header,
	}, s.maxHeaderSize)
	if err != nil {
		return
	}

	// Call the closed hook once closed.
	go func() {
		select {
		case <-stream.ClosedChan():
		case <-s.ClosingChan():
		}
		s.handler.hookOnStreamClosed(cctx, id, nil)
	}()

	// Create our typed stream.
	ts = newTypedRWStream(stream, s.codec, maxArgSize, maxRetSize)
	return
}

// TODO: 2020/07/16 skaldesh: OpenRawStream?
func (s *session) OpenStream(ctx context.Context, id string) (stream transport.Stream, err error) {
	// Create a new client context.
	cctx := newContext(ctx, s)

	// Call the OnStream hooks.
	err = s.handler.hookOnStream(cctx, id)
	if err != nil {
		return
	}

	// Call the closed hook if an error occurs.
	defer func() {
		if err != nil {
			s.handler.hookOnStreamClosed(cctx, id, err)
		}
	}()

	// Open the stream.
	stream, err = s.openStream(ctx, api.StreamTypeRaw, &api.StreamRaw{
		ID:   id,
		Data: cctx.header,
	}, s.maxHeaderSize)
	if err != nil {
		return
	}

	// Call the closed hook once closed.
	go func() {
		select {
		case <-stream.ClosedChan():
		case <-s.ClosingChan():
		}
		s.handler.hookOnStreamClosed(cctx, id, nil)
	}()
	return
}

func (s *session) openStream(
	ctx context.Context,
	streamType api.StreamType,
	header interface{},
	maxHeaderSize int,
) (
	stream transport.Stream,
	err error,
) {
	// Open the stream through our conn.
	stream, err = s.conn.OpenStream(ctx)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = stream.Close()
		}
	}()

	// Set a write deadline, if needed.
	if deadline, ok := ctx.Deadline(); ok {
		err = stream.SetWriteDeadline(deadline)
		if err != nil {
			return
		}
	}

	// Write the stream type.
	_, err = stream.Write([]byte{byte(streamType)})
	if err != nil {
		return
	}

	// Write the header if set.
	if header != nil {
		err = packet.WriteEncode(stream, &header, api.Codec, maxHeaderSize)
		if err != nil {
			return
		}
	}

	// Reset the deadlines.
	err = stream.SetDeadline(time.Time{})
	return
}
