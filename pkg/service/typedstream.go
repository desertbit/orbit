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
	"fmt"
	"io"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/packet"
	"github.com/desertbit/orbit/pkg/transport"
)

const (
	maxTypedStreamErrorSize = 4096 // 4 KB
)

type TypedRStream interface {
	Read(data interface{}) error
}

type TypedWStream interface {
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

func (s *typedRWStream) IsClosed() bool {
	return s.stream.IsClosed()
}

func (s *typedRWStream) ClosedChan() <-chan struct{} {
	return s.stream.ClosedChan()
}

func (s *typedRWStream) Read(data interface{}) (err error) {
	// Read first the type off of the wire.
	t := make([]byte, 1)
	for {
		n, err := s.stream.Read(t)
		if err != nil {
			return s.checkClosedErr(err)
		} else if n == 1 {
			break
		}
	}

	switch t[0] {
	case byte(api.TypedStreamTypeData):
		// Read the data packet.
		err = packet.ReadDecode(s.stream, &data, s.codec, s.maxRetSize)
		if err != nil {
			return s.checkClosedErr(err)
		}
	case byte(api.TypedStreamTypeError):
		// Read the error packet.
		var tErr api.TypedStreamError
		err = packet.ReadDecode(s.stream, &tErr, s.codec, maxTypedStreamErrorSize)
		if err != nil {
			return s.checkClosedErr(err)
		}

		// Build our error.
		err = Err(errors.New(tErr.Err), tErr.Err, tErr.Code)
		// Close the stream.
		_ = s.stream.Close()
	default:
		return fmt.Errorf("unknown typed stream type: %v", t[0])
	}

	return
}

func (s *typedRWStream) Write(data interface{}) (err error) {
	// Write first the type on the wire.
	n, err := s.stream.Write([]byte{byte(api.TypedStreamTypeData)})
	if err != nil {
		return s.checkClosedErr(err)
	} else if n != 1 {
		return io.ErrShortWrite
	}

	// Now write the data packet.
	err = packet.WriteEncode(s.stream, data, s.codec, s.maxArgSize)
	if err != nil {
		return s.checkClosedErr(err)
	}
	return
}

func (s *typedRWStream) closeWithErr(rErr Error) (err error) {
	if rErr == nil {
		return
	}

	// Write first the type on the wire.
	n, err := s.stream.Write([]byte{byte(api.TypedStreamTypeError)})
	if err != nil {
		return s.checkClosedErr(err)
	} else if n != 1 {
		return io.ErrShortWrite
	}

	// Now write the error packet.
	err = packet.WriteEncode(s.stream, api.TypedStreamError{Err: rErr.Msg(), Code: rErr.Code()}, s.codec, s.maxArgSize)
	if err != nil {
		return s.checkClosedErr(err)
	}
}

func (s *typedRWStream) checkClosedErr(err error) error {
	if errors.Is(err, io.EOF) || s.stream.IsClosed() {
		return ErrClosed
	}
	return err
}
