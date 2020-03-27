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
	"net"
	"sync"
	"testing"
	"time"

	"github.com/desertbit/orbit/pkg/codec/msgpack"
	"github.com/desertbit/orbit/pkg/packet"
	"github.com/stretchr/testify/assert"
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

	// Start a routine that reads from the connection.
	// Needed, since net.Pipe() returns conn with no internal buffering.
	done := make(chan struct{})
	go func() {
		err := packet.ReadDecode(connR, &ret, msgpack.Codec)
		if err != nil {
			t.Fatal(err)
		}
		done <- struct{}{}
	}()

	err := packet.WriteEncode(connW, data, msgpack.Codec)
	require.NoError(t, err)

	// Wait for the read to finish.
	select {
	case <-done:
	case <-time.After(time.Millisecond * 10):
		t.Fatal("timeout")
	}

	require.Equal(t, data.Msg, ret.Msg)
}

func TestWrite(t *testing.T) {
	t.Parallel()

	data := []byte("This is some test data")

	cases := []struct {
		data        []byte
		buffer      []byte
		shouldPanic bool
	}{
		{ // 0
			data:   data,
			buffer: nil,
		},
		{
			data:   data,
			buffer: make([]byte, 1),
		},
		{
			data:   data,
			buffer: make([]byte, len(data)),
		},
		{
			data:   data,
			buffer: make([]byte, len(data)+1),
		},
		{
			data:   []byte{},
			buffer: nil,
		},
		{ // 5
			data:   nil,
			buffer: nil,
		},
		{
			data:        make([]byte, packet.MaxSize+1),
			buffer:      nil,
			shouldPanic: true,
		},
	}

	connW, connR := net.Pipe()

	// Start a routine that reads from the connection.
	// Needed, since net.Pipe() returns conn with no internal buffering.
	done := make(chan struct{})
	go func() {
		for i, c := range cases {
			func() {
				defer func() { done <- struct{}{} }()
				if c.shouldPanic {
					return
				}

				// Try to read the same data off of the buffer.
				ret, err := packet.Read(connR, nil)
				if len(c.data) == 0 {
					if !assert.Error(t, packet.ErrZeroData, "read case %d", i) {
						return
					}
				} else {
					if !assert.NoErrorf(t, err, "read case %d", i) {
						return
					}
					if !assert.Equalf(t, c.data, ret, "read case %d", i) {
						return
					}
				}
			}()
		}
	}()

	for i, c := range cases {
		if c.shouldPanic {
			require.Panicsf(t, func() {
				_ = packet.Write(connW, c.data)
			}, "write case %d", i)
		} else {
			err := packet.Write(connW, c.data)
			require.NoErrorf(t, err, "write case %d", i)
		}

		// Wait for read to finish.
		select {
		case <-done:
		case <-time.After(time.Millisecond * 100):
			if t.Failed() {
				return
			}
			t.Fatal("timeout")
		}
	}
}

func TestReadDecode(t *testing.T) {
	t.Parallel()

	type test struct {
		Msg string
	}

	data := test{Msg: "hello world"}

	connW, connR := net.Pipe()

	// Write in a goroutine, because net.Pipe() does not buffer.
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := packet.WriteEncode(connW, data, msgpack.Codec)
		require.NoError(t, err)
		wg.Done()
	}()

	var ret test
	err := packet.ReadDecode(connR, &ret, msgpack.Codec)
	require.NoError(t, err)
	require.Equal(t, data.Msg, ret.Msg)
	wg.Wait()
}

func TestRead(t *testing.T) {
	t.Parallel()

	data := []byte("This is some test data")

	cases := []struct {
		data        []byte
		buffer      []byte
		reuseBuffer bool
	}{
		{
			data:   data,
			buffer: nil,
		},
		{
			data:   data,
			buffer: make([]byte, 1),
		},
		{
			data:        data,
			reuseBuffer: true,
			buffer:      make([]byte, len(data)),
		},
		{
			data:        data,
			reuseBuffer: true,
			buffer:      make([]byte, len(data)+1),
		},
		{
			data:   []byte{},
			buffer: nil,
		},
		{
			data:   []byte{},
			buffer: nil,
		},
	}

	connW, connR := net.Pipe()

	// Start a routine that reads from the connection.
	// Needed, since net.Pipe() returns conn with no internal buffering.
	done := make(chan struct{})
	go func() {
		for i, c := range cases {
			func() {
				defer func() { done <- struct{}{} }()

				// Try to read the same data off of the buffer.
				ret, err := packet.Read(connR, c.buffer)
				if len(c.data) == 0 {
					if !assert.Error(t, packet.ErrZeroData, "read case %d", i) {
						return
					}
				} else {
					if !assert.NoErrorf(t, err, "read case %d", i) {
						return
					}
					if !assert.Equalf(t, c.data, ret, "read case %d", i) {
						return
					}
					if c.reuseBuffer && !assert.Samef(t, &c.buffer[0], &ret[0], "read case %d", i) {
						return
					}
				}
			}()
		}
	}()

	for i, c := range cases {
		err := packet.Write(connW, c.data)
		require.NoErrorf(t, err, "write case %d", i)

		// Wait for read to finish.
		select {
		case <-done:
		case <-time.After(time.Millisecond * 100):
			if t.Failed() {
				return
			}
			t.Fatal("timeout")
		}
	}
}
