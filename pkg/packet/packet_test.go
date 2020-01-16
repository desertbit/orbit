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
	"sync"
	"testing"
	"time"

	"github.com/desertbit/orbit/pkg/codec/msgpack"
	"github.com/desertbit/orbit/pkg/packet"
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
		err := ReadDecode(connR, &ret, msgpack.Codec, 16384, 50*time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		done <- struct{}{}
	}()

	err := WriteEncode(connW, data, msgpack.Codec, 16384, 50*time.Millisecond)
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

	err := packet.WriteTimeout(connW, data, 16384, 50*time.Millisecond)
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
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := WriteEncode(connW, data, msgpack.Codec, 16384, 50*time.Millisecond)
		require.NoError(t, err)
		wg.Done()
	}()

	var ret test
	err := ReadDecode(connR, &ret, msgpack.Codec, 16384, 50*time.Millisecond)
	require.NoError(t, err)
	require.Equal(t, data.Msg, ret.Msg)
	wg.Wait()
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
	_, err := packet.ReadTimeout(connR, nil, 16384, time.Nanosecond)
	if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
		t.Fatalf("expected timeout error, but error was '%v'", err)
	}

	// Read normally.
	ret, err := packet.ReadTimeout(connR, nil, 16384, 50*time.Millisecond)
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

func TestStream(t *testing.T) {
	t.Parallel()

	connW, connR := net.Pipe()

	testData := []byte("This is some test data")
	dst := &bytes.Buffer{}

	// Start a writing goroutine.
	go func() {
		// Write some normal test data.
		err := packet.WriteStream(connW, bytes.NewReader(testData))
		require.NoError(t, err)

		// Write some normal test data again.
		err = packet.WriteStream(connW, bytes.NewReader(testData))
		require.NoError(t, err)

		// Now close the connection and try again.
		require.NoError(t, connW.Close())
		err = packet.WriteStream(connW, bytes.NewReader(testData))
		require.Error(t, err)
	}()

	// Read the normal test data.
	err := packet.ReadStream(connR, dst)
	require.NoError(t, err)
	require.Equal(t, testData, dst.Bytes())
	dst.Reset()

	// Try a destination writer that is closed.
	dstClosed, _ := net.Pipe()
	require.NoError(t, dstClosed.Close())
	err = packet.ReadStream(connR, dstClosed)
	require.Error(t, err)

	// Read the data from the closed connection. There should be no error
	// as the other side closed the connection.
	err = packet.ReadStream(connR, dst)
	require.NoError(t, err)
}

func TestStreamTimeout(t *testing.T) {
	t.Parallel()

	connW, connR := net.Pipe()

	testData := []byte("This is some test data")
	dst := &bytes.Buffer{}

	// Start a writing goroutine.
	continueChan := make(chan struct{})
	go func() {
		// Write some normal test data without a timeout.
		buf := make([]byte, 8)
		err := packet.WriteStreamBufTimeout(connW, bytes.NewReader(testData), buf, 0)
		require.NoError(t, err)

		// Write some normal test data without a timeout and buffer.
		err = packet.WriteStreamBufTimeout(connW, bytes.NewReader(testData), nil, -1)
		require.NoError(t, err)

		// Wait for a continue signal.
		<-continueChan

		// Write with a timeout.
		err = packet.WriteStreamBufTimeout(connW, bytes.NewReader(testData), nil, 1)
		require.Error(t, err)
		require.Implements(t, (*net.Error)(nil), err)
		require.True(t, (err.(net.Error)).Timeout())

		// Signal the read routine to continue.
		continueChan <- struct{}{}
	}()

	// Read the normal test data without a timeout.
	buf := make([]byte, 8)
	err := packet.ReadStreamBufTimeout(connR, dst, buf, 0)
	require.NoError(t, err)
	require.Equal(t, testData, dst.Bytes())
	dst.Reset()

	// Read the normal test data without a timeout and buffer.
	err = packet.ReadStreamBufTimeout(connR, dst, nil, -1)
	require.NoError(t, err)
	require.Equal(t, testData, dst.Bytes())
	dst.Reset()

	// Read with a timeout.
	err = packet.ReadStreamBufTimeout(connR, dst, nil, 1)
	require.Error(t, err)
	require.Implements(t, (*net.Error)(nil), err)
	require.True(t, (err.(net.Error)).Timeout())

	// Signal the write routine to continue.
	continueChan <- struct{}{}

	// Wait for a continue signal.
	<-continueChan
}
