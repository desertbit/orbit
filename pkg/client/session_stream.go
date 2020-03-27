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

func (s *session) OpenStream(ctx context.Context, id string) (stream transport.Stream, err error) {
	stream, err = s.openStream(ctx, id, api.StreamTypeRaw)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = stream.Close()
		}
	}()

	// Create a new client context.
	cctx := newContext(ctx, s)

	// Call the OnStream hooks.
	err = s.handler.hookOnStream(cctx, id, stream)
	if err != nil {
		return nil, err
	}
	return
}

func (s *session) openStream(ctx context.Context, id string, t api.StreamType) (stream transport.Stream, err error) {
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

	// Create the initial data that signals to the remote peer,
	// which stream we want to open.
	data := api.InitStream{
		ID:   id,
		Type: t,
	}

	// Set a write deadline, if needed.
	if deadline, ok := ctx.Deadline(); ok {
		err = stream.SetWriteDeadline(deadline)
		if err != nil {
			return
		}
	}

	// Write the initial request to the stream.
	err = packet.WriteEncode(stream, &data, api.Codec)
	if err != nil {
		return
	}

	// Reset the deadlines.
	err = stream.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	return
}
