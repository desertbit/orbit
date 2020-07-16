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
	wOnly      bool
}

func newTypedRWStream(s transport.Stream, cc codec.Codec, mas, mrs int, wOnly bool) *typedRWStream {
	return &typedRWStream{stream: s, codec: cc, maxArgSize: mas, maxRetSize: mrs, wOnly: wOnly}
}

func (s *typedRWStream) IsClosed() bool {
	return s.stream.IsClosed()
}

func (s *typedRWStream) ClosedChan() <-chan struct{} {
	return s.stream.ClosedChan()
}

func (s *typedRWStream) Read(data interface{}) (err error) {
	// Read first the type off of the wire.
	ts, err := s.readTypedStreamType()
	if err != nil {
		return s.checkErr(err)
	}

	switch ts {
	case api.TypedStreamTypeData:
		// Read the data packet.
		err = packet.ReadDecode(s.stream, &data, s.codec, s.maxRetSize)
		if err != nil {
			return s.checkErr(err)
		}
	case api.TypedStreamTypeError:
		// Read the error packet.
		var tErr api.TypedStreamError
		err = packet.ReadDecode(s.stream, &tErr, api.Codec, maxTypedStreamErrorSize)
		if err != nil {
			return s.checkErr(err)
		}

		// Build our error.
		err = NewErr(errors.New(tErr.Err), tErr.Err, tErr.Code)
		// Close the stream.
		_ = s.stream.Close()
	default:
		return fmt.Errorf("unknown typed stream type: %v", t[0])
	}

	return
}

func (s *typedRWStream) Write(data interface{}) (err error) {
	// Write first the type on the wire.
	_, err = s.stream.Write([]byte{byte(api.TypedStreamTypeData)})
	if err != nil {
		// If the stream is closed, check for an error sent by the client.
		if s.wOnly && s.isClosedErr(err) {
			err = s.readErr()
		}
		return s.checkErr(err)
	}

	// Now write the data packet.
	err = packet.WriteEncode(s.stream, data, s.codec, s.maxArgSize)
	if err != nil {
		// If the stream is closed, check for an error sent by the client.
		if s.wOnly && s.isClosedErr(err) {
			err = s.readErr()
		}
		return s.checkErr(err)
	}
	return
}

//###############//
//### Private ###//
//###############//

func (s *typedRWStream) readTypedStreamType() (ts api.TypedStreamType, err error) {
	// Read first the type off of the wire.
	var (
		n int
		t = make([]byte, 1)
	)
	for {
		n, err = s.stream.Read(t)
		if err != nil {
			return
		} else if n == 1 {
			break
		}
	}
	ts = api.TypedStreamType(t[0])
	return
}

func (s *typedRWStream) readErr() (err error) {
	ts, err := s.readTypedStreamType()
	if err != nil || ts != api.TypedStreamTypeError {
		return
	}

	// Read the error packet.
	var tErr api.TypedStreamError
	err = packet.ReadDecode(s.stream, &tErr, api.Codec, maxTypedStreamErrorSize)
	if err != nil {
		return
	}

	// Build our error.
	return NewErr(errors.New(tErr.Err), tErr.Err, tErr.Code)
}

func (s *typedRWStream) closeWithErr(rErr Error) (err error) {
	defer s.stream.Close()

	if rErr == nil {
		return
	}

	// Write first the type on the wire.
	_, err = s.stream.Write([]byte{byte(api.TypedStreamTypeError)})
	if err != nil {
		return s.checkErr(err)
	}

	// Now write the error packet.
	err = packet.WriteEncode(
		s.stream,
		api.TypedStreamError{Err: rErr.Msg(), Code: rErr.Code()},
		api.Codec,
		s.maxArgSize,
	)
	if err != nil {
		return s.checkErr(err)
	}
	return
}

func (s *typedRWStream) checkErr(err error) error {
	if s.isClosedErr(err) {
		return ErrClosed
	}
	return err
}

func (s *typedRWStream) isClosedErr(err error) bool {
	return errors.Is(err, io.EOF) || s.stream.IsClosed()
}
