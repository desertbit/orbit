/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2018 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2018 Sebastian Borchers <sebastian[at]desertbit.com>
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

package packet_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/desertbit/orbit/pkg/codec/msgpack"
	"github.com/stretchr/testify/require"
)

func TestWriteEncode(t *testing.T) {
	t.Parallel()

	type test struct {
		Msg string
	}

	data := test{Msg: "hello world"}
	var ret test

	connW, connR := net.Pipe()
	defer func() {
		_ = connW.Close()
		_ = connR.Close()
	}()
	// Start a routine that reads from the connection.
	// Needed, since net.Pipe() returns conn with no internal buffering.
	done := make(chan struct{})
	go func() {
		err := ReadDecode(connR, &ret, msgpack.Codec, 16384, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		done <- struct{}{}
	}()

	err := WriteEncode(connW, data, msgpack.Codec, 16384, time.Second)
	require.NoError(t, err)

	// Wait for the read to finish.
	select {
	case <-done:
	case <-time.After(time.Millisecond * 10):
		t.Fatal("timeout")
	}

	require.Equal(t, data.Msg, ret.Msg)
}

func TestWriteTimeout(t *testing.T) {
	t.Parallel()

	data := []byte("This is some test data")
	var ret []byte

	connW, connR := net.Pipe()
	defer func() {
		_ = connW.Close()
		_ = connR.Close()
	}()
	// Start a routine that reads from the connection.
	// Needed, since net.Pipe() returns conn with no internal buffering.
	done := make(chan struct{})
	go func() {
		// Try to read the same data off of the conn.
		var err error
		ret, err = Read(connR, nil, 16384)
		if err != nil {
			t.Fatal(err)
		}

		done <- struct{}{}
	}()

	err := WriteTimeout(connW, data, 16384, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Wait, until the read has finished.
	select {
	case <-done:
	case <-time.After(time.Millisecond * 10):
		t.Fatal("timeout")
	}

	require.Equal(t, data, ret)
}

func TestWrite(t *testing.T) {
	t.Parallel()

	data := []byte("This is some test data")

	cases := []struct {
		data           []byte
		maxPayloadSize int
		buffer         []byte
		shouldFail     bool
		shouldPanic    bool
	}{
		{
			data:           data,
			maxPayloadSize: len(data),
			buffer:         nil,
		},
		{
			data:           data,
			maxPayloadSize: len(data),
			buffer:         make([]byte, 1),
		},
		{
			data:           data,
			maxPayloadSize: len(data),
			buffer:         make([]byte, len(data)),
		},
		{
			data:           data,
			maxPayloadSize: len(data),
			buffer:         make([]byte, len(data)+1),
		},
		{
			data:           data,
			maxPayloadSize: 1,
			buffer:         nil,
			shouldFail:     true,
		},
		{
			data:           data,
			maxPayloadSize: 0,
			buffer:         nil,
			shouldFail:     true,
		},
		{
			data:           []byte{},
			maxPayloadSize: 1,
			buffer:         nil,
		},
		{
			data:           []byte{},
			maxPayloadSize: 0,
			buffer:         nil,
		},
		{
			data:           []byte{},
			maxPayloadSize: -1,
			buffer:         nil,
			shouldPanic:    true,
		},
	}

	for i, c := range cases {
		buff := bytes.Buffer{}
		if c.shouldPanic {
			require.Panicsf(t, func() {
				_ = Write(&buff, c.data, c.maxPayloadSize)
			}, "case %d", i+1)
			continue
		}

		err := Write(&buff, c.data, c.maxPayloadSize)
		if c.shouldFail {
			require.Errorf(t, err, "case %d", i+1)
			continue
		}

		require.NoErrorf(t, err, "case %d", i+1)

		// Try to read the same data off of the buffer.
		ret, err := Read(&buff, nil, 16384)
		require.NoErrorf(t, err, "case %d", i+1)
		require.Equalf(t, c.data, ret, "case %d", i+1)
	}
}

func TestReadDecode(t *testing.T) {
	t.Parallel()

	type test struct {
		Msg string
	}

	data := test{Msg: "hello world"}

	connW, connR := net.Pipe()
	defer func() {
		_ = connW.Close()
		_ = connR.Close()
	}()

	// Write in a goroutine, because net.Pipe() does not buffer.
	go func() {
		err := WriteEncode(connW, data, msgpack.Codec, 16384, time.Second)
		require.NoError(t, err)
	}()

	var ret test
	err := ReadDecode(connR, &ret, msgpack.Codec, 16384, time.Second)
	require.NoError(t, err)
	require.Equal(t, data.Msg, ret.Msg)
}

func TestReadTimeout(t *testing.T) {
	t.Parallel()

	data := []byte("This is some test data")

	connW, connR := net.Pipe()
	defer func() {
		_ = connW.Close()
		_ = connR.Close()
	}()

	// Write in a goroutine, because net.Pipe() does not buffer.
	go func() {
		err := Write(connW, data, 16384)
		require.NoError(t, err)
	}()

	// Read with a timeout.
	_, err := ReadTimeout(connR, nil, 16384, time.Nanosecond)
	if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
		t.Fatalf("expected timeout error, but error was '%v'", err)
	}

	// Read normally.
	ret, err := ReadTimeout(connR, nil, 16384, time.Second)
	require.NoError(t, err)
	require.Equal(t, data, ret)
}

func TestRead(t *testing.T) {
	t.Parallel()

	data := []byte("This is some test data")

	cases := []struct {
		data           []byte
		maxPayloadSize int
		buffer         []byte
		shouldFail     bool
		shouldPanic    bool
	}{
		{
			data:           data,
			maxPayloadSize: len(data),
			buffer:         nil,
		},
		{
			data:           data,
			maxPayloadSize: len(data),
			buffer:         make([]byte, 1),
		},
		{
			data:           data,
			maxPayloadSize: len(data),
			buffer:         make([]byte, len(data)),
		},
		{
			data:           data,
			maxPayloadSize: len(data),
			buffer:         make([]byte, len(data)+1),
		},
		{
			data:           data,
			maxPayloadSize: 1,
			buffer:         nil,
			shouldFail:     true,
		},
		{
			data:           data,
			maxPayloadSize: 0,
			buffer:         nil,
			shouldFail:     true,
		},
		{
			data:           []byte{},
			maxPayloadSize: 1,
			buffer:         nil,
		},
		{
			data:           []byte{},
			maxPayloadSize: 0,
			buffer:         nil,
		},
		{
			data:           []byte{},
			maxPayloadSize: -1,
			buffer:         nil,
			shouldPanic:    true,
		},
	}

	for i, c := range cases {
		buff := bytes.Buffer{}

		err := Write(&buff, c.data, 16384)
		require.NoErrorf(t, err, "case %d", i+1)

		// Try to read the same data off of the buffer.

		if c.shouldPanic {
			require.Panicsf(t, func() {
				_, _ = Read(&buff, c.buffer, c.maxPayloadSize)
			}, "case %d", i+1)
			continue
		}

		ret, err := Read(&buff, c.buffer, c.maxPayloadSize)
		if c.shouldFail {
			require.Errorf(t, err, "case %d", i+1)
			continue
		}

		require.NoErrorf(t, err, "case %d", i+1)
		require.Equalf(t, c.data, ret, "case %d", i+1)
	}
}
