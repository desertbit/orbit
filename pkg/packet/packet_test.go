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

package packet_test

import (
	"net"
	"testing"
	"time"

	"github.com/desertbit/orbit/pkg/packet"
	"github.com/stretchr/testify/require"
)

/*
func TestReadWriteEncode(t *testing.T) {
	t.Parallel()

	type test struct {
		Msg string
	}

	data := test{Msg: "hello world"}
	var ret test

	connW, connR := net.Pipe()
	defer connW.Close()
	defer connR.Close()

	// Start a routine that reads from the connection.
	// Needed, since net.Pipe() returns conn with no internal buffering.
	errChan := make(chan error)
	go func() {
		errChan <- packet.ReadDecode(connR, &ret, msgpack.Codec, 0)
	}()

	err := packet.WriteEncode(connW, data, msgpack.Codec, 0)
	require.NoError(t, err)

	// Wait for the read to finish.
	select {
	case err = <-errChan:
		require.NoError(t, err)
	case <-time.After(time.Millisecond * 100):
		t.Fatal("timeout")
	}

	require.Equal(t, data.Msg, ret.Msg)
}

func TestWrite(t *testing.T) {
	t.Parallel()

	data := []byte("This is some test data")

	cases := []struct {
		data           []byte
		buffer         []byte
		maxPayloadSize int
		shouldFail     bool
		exactError     error
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
		{
			data:   nil,
			buffer: nil,
		},
		{
			data:       make([]byte, packet.MaxSize+1),
			buffer:     nil,
			shouldFail: true,
			exactError: packet.ErrMaxPayloadSizeExceeded,
		},
		{
			data:           data,
			maxPayloadSize: len(data),
		},
		{
			data:           data,
			maxPayloadSize: len(data) + 1,
		},
		{
			data:           data,
			maxPayloadSize: len(data) - 1,
			shouldFail:     true,
			exactError:     packet.ErrMaxPayloadSizeExceeded,
		},
		{
			data:           data,
			maxPayloadSize: packet.MaxSize,
		},
		{
			data:           data,
			maxPayloadSize: packet.MaxSize + 1,
		},
		{
			data:           make([]byte, packet.MaxSize),
			maxPayloadSize: packet.MaxSize,
		},
		{
			data:           make([]byte, packet.MaxSize+1),
			maxPayloadSize: packet.MaxSize,
			shouldFail:     true,
			exactError:     packet.ErrMaxPayloadSizeExceeded,
		},
		{
			data:           make([]byte, packet.MaxSize+1),
			maxPayloadSize: packet.MaxSize + 1,
			shouldFail:     true,
			exactError:     packet.ErrMaxPayloadSizeExceeded,
		},
		{
			data:       make([]byte, packet.MaxSize+1),
			shouldFail: true,
			exactError: packet.ErrMaxPayloadSizeExceeded,
		},
	}

	connW, connR := net.Pipe()
	defer connW.Close()
	defer connR.Close()

	// Start a routine that reads from the connection.
	// Needed, since net.Pipe() returns conn with no internal buffering.
	go func() {
		buf := make([]byte, 1000)
		for {
			_, err := connR.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	for i, c := range cases {
		err := packet.Write(connW, c.data, c.maxPayloadSize)
		if c.shouldFail {
			require.Errorf(t, err, "write case %d", i)
			if c.exactError != nil {
				require.Equalf(t, err, c.exactError, "write case %d", i)
			}
		} else {
			require.NoErrorf(t, err, "write case %d", i)
		}
	}
}
*/
func TestRead(t *testing.T) {
	t.Parallel()

	data := []byte("This is some test data")

	cases := []struct {
		data           []byte
		buffer         []byte
		maxPayloadSize int
		reuseBuffer    bool
		shouldFail     bool
		exactError     error
		flushReadBytes int
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
			data:       []byte{},
			buffer:     nil,
			shouldFail: true,
			exactError: packet.ErrZeroData,
		},
		{
			data:       nil,
			buffer:     nil,
			shouldFail: true,
			exactError: packet.ErrZeroData,
		},
		{
			data:           data,
			maxPayloadSize: len(data),
		},
		{
			data:           data,
			maxPayloadSize: len(data) + 1,
		},
		{
			data:           data,
			maxPayloadSize: len(data) - 1,
			shouldFail:     true,
			exactError:     packet.ErrMaxPayloadSizeExceeded,
			flushReadBytes: len(data),
		},
	}

	connW, connR := net.Pipe()
	defer connW.Close()
	defer connR.Close()

	// Start a routine that reads from the connection.
	// Needed, since net.Pipe() returns conn with no internal buffering.
	errChan := make(chan error, 1)
	go func() {
		for _, c := range cases {
			errChan <- packet.Write(connW, c.data, 0)
		}
	}()

	for i, c := range cases {
		// Try to read the same data off of the buffer.
		ret, err := packet.Read(connR, c.buffer, c.maxPayloadSize)
		if err != nil {
			if c.shouldFail {
				require.Errorf(t, err, "read case %d", i)
				if c.exactError != nil {
					require.Equalf(t, err, c.exactError, "read case %d", i)
				}
			} else {
				require.NoErrorf(t, err, "read case %d", i)
			}
		} else {
			require.Equalf(t, c.data, ret, "read case %d", i)
			if c.reuseBuffer {
				require.Samef(t, &c.buffer[0], &ret[0], "read case %d", i)
			}
		}

		if c.flushReadBytes > 0 {
			buf := make([]byte, c.flushReadBytes)
			var bytesRead int
			for bytesRead < c.flushReadBytes {
				n, err := connR.Read(buf[bytesRead:])
				require.NoErrorf(t, err, "read case %d", i)
				bytesRead += n
			}
		}

		select {
		case err := <-errChan:
			require.NoErrorf(t, err, "write case %d", i)
		case <-time.After(time.Millisecond * 500):
			t.Fatal("timeout", i)
		}
	}
}
