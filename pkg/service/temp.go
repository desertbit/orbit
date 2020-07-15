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

package service

import (
	"errors"
	"io"

	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/packet"
	"github.com/desertbit/orbit/pkg/transport"
)

type typedBiStream struct {
	stream  transport.Stream
	codec   codec.Codec
	maxSize int
}

func newBidirectionalStream(s transport.Stream, cc codec.Codec, ms int) *typedBiStream {
	return &typedBiStream{stream: s, codec: cc, maxSize: ms}
}

func (v1 *typedBiStream) Read(data interface{}) (err error) {
	err = packet.ReadDecode(v1.stream, &data, v1.codec, v1.maxSize)
	if err != nil {
		if errors.Is(err, packet.ErrZeroData) || errors.Is(err, io.EOF) || v1.stream.IsClosed() {
			err = ErrClosed
		}
		return
	}
	return
}

func (v1 *typedBiStream) Write(data interface{}) (err error) {
	err = packet.WriteEncode(v1.stream, data, v1.codec, v1.maxSize)
	if err != nil {
		if errors.Is(err, io.EOF) || v1.stream.IsClosed() {
			return ErrClosed
		}
	}
	return
}
