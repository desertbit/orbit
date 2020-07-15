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
	"errors"
	"io"

	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/packet"
	"github.com/desertbit/orbit/pkg/transport"
)

type TypedStreamCloser interface {
	Close() error
}

type TypedRStream interface {
	TypedStreamCloser
	Read(data interface{}) error
}

type TypedWStream interface {
	TypedStreamCloser
	Write(data interface{}) error
}

type TypedRWStream interface {
	TypedRStream
	TypedWStream
}

type typedRWStream struct {
	stream     transport.Stream
	codec      codec.Codec
	maxArgSize int
	maxRetSize int
}

func newTypedRWStream(s transport.Stream, cc codec.Codec, mas, mrs int) *typedRWStream {
	return &typedRWStream{stream: s, codec: cc, maxArgSize: mas, maxRetSize: mrs}
}

func (s *typedRWStream) Close() error {
	return s.stream.Close()
}

func (s *typedRWStream) Read(data interface{}) (err error) {
	err = packet.ReadDecode(s.stream, &data, s.codec, s.maxRetSize)
	if err != nil {
		if errors.Is(err, io.EOF) || s.stream.IsClosed() {
			err = ErrClosed
		}
		return
	}
	return
}

func (s *typedRWStream) Write(data interface{}) (err error) {
	err = packet.WriteEncode(s.stream, data, s.codec, s.maxArgSize)
	if err != nil {
		if errors.Is(err, io.EOF) || s.stream.IsClosed() {
			return ErrClosed
		}
	}
	return
}
