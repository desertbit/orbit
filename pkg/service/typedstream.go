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

type TypedStreamCloser interface {
	ClosedChan() <-chan struct{}
	IsClosed() bool
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
	stream       transport.Stream
	codec        codec.Codec
	maxReadSize  int
	maxWriteSize int
	wOnly        bool
}

func newTypedRWStream(s transport.Stream, cc codec.Codec, maxReadSize, maxWriteSize int, wOnly bool) *typedRWStream {
	return &typedRWStream{
		stream:       s,
		codec:        cc,
		maxReadSize:  maxReadSize,
		maxWriteSize: maxWriteSize,
		wOnly:        wOnly,
	}
}

func (s *typedRWStream) IsClosed() bool {
	return s.stream.IsClosed()
}

func (s *typedRWStream) ClosedChan() <-chan struct{} {
	return s.stream.ClosedChan()
}

// Returns ErrClosed, Error or error.
func (s *typedRWStream) Read(data interface{}) (err error) {
	// Read first the type off the wire.
	ts, err := s.readTypedStreamType()
	if err != nil {
		return s.checkErr(err)
	}

	switch ts {
	case api.TypedStreamTypeData:
		// Read the data packet.
		err = packet.ReadDecode(s.stream, &data, s.codec, s.maxReadSize)
		if err != nil {
			return s.checkErr(err)
		}
	case api.TypedStreamTypeError:
		// Close the stream in any case now.
		defer s.stream.Close()

		// Read the error packet.
		var tErr api.TypedStreamError
		err = packet.ReadDecode(s.stream, &tErr, api.Codec, maxTypedStreamErrorSize)
		if err != nil {
			return s.checkErr(err)
		}

		// Build our error.
		err = NewError(errors.New(tErr.Err), tErr.Err, tErr.Code)
	default:
		return fmt.Errorf("unknown typed stream type: %v", ts)
	}

	return
}

// Returns ErrClosed, Error or error.
func (s *typedRWStream) Write(data interface{}) (err error) {
	// Write first the type on the wire.
	_, err = s.stream.Write([]byte{byte(api.TypedStreamTypeData)})
	if err != nil {
		// If the stream is closed, check for an error sent by the client.
		return s.checkErr(s.checkWriteErr(err))
	}

	// Now write the data packet.
	err = packet.WriteEncode(s.stream, data, s.codec, s.maxWriteSize)
	if err != nil {
		// If the stream is closed, check for an error sent by the client.
		return s.checkErr(s.checkWriteErr(err))
	}
	return
}

//###############//
//### Private ###//
//###############//

func (s *typedRWStream) readTypedStreamType() (ts api.TypedStreamType, err error) {
	// Read first the type off the wire.
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

func (s *typedRWStream) closeWithErr(sErr api.TypedStreamError) (err error) {
	defer s.stream.Close()

	// Write first the type on the wire.
	_, err = s.stream.Write([]byte{byte(api.TypedStreamTypeError)})
	if err != nil {
		return s.checkErr(err)
	}

	// Now write the error packet.
	err = packet.WriteEncode(s.stream, sErr, api.Codec, s.maxReadSize)
	if err != nil {
		return s.checkErr(err)
	}
	return
}

func (s *typedRWStream) checkWriteErr(err error) error {
	// Only check for write errors in write-only mode and only for closed errors.
	if !s.wOnly || !s.isClosedErr(err) {
		return err
	}

	// Try to read the typed stream type of the wire.
	var rErr error
	ts, rErr := s.readTypedStreamType()
	if rErr != nil || ts != api.TypedStreamTypeError {
		return err
	}

	// Read the error packet.
	var tErr api.TypedStreamError
	rErr = packet.ReadDecode(s.stream, &tErr, api.Codec, maxTypedStreamErrorSize)
	if rErr != nil {
		return err
	}

	// Always prefer to return our custom error.
	return NewError(errors.New(tErr.Err), tErr.Err, tErr.Code)
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
