/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018 Sebastian Borchers <sebastian[at].desertbit.com>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package packet_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/desertbit/orbit/codec/msgpack"
	"github.com/desertbit/orbit/packet"
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
		err := packet.ReadDecode(connR, &ret, msgpack.Codec, 16384, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		done <- struct{}{}
	}()

	err := packet.WriteEncode(connW, data, msgpack.Codec, 16384, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for the read to finish
	<-done

	if ret.Msg != data.Msg {
		t.Fatalf("wrong return msg; expected '%s', got '%s'", data.Msg, ret.Msg)
	}
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
		ret, err = packet.Read(connR, nil, 16384)
		if err != nil {
			t.Fatal(err)
		}

		done <- struct{}{}
	}()

	err := packet.WriteTimeout(connW, data, 16384, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Wait, until the read has finished
	<-done

	if !bytes.Equal(data, ret) {
		t.Fatalf("read data was not equal to sent data, expected %v, got %v", data, ret)
	}
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
		func() {
			defer func() {
				e := recover()
				if e == nil && c.shouldPanic {
					t.Fatalf("expected a panic for case %d", i+1)
				} else if e != nil && !c.shouldPanic {
					t.Fatalf("panic for case %d", i+1)
				}
			}()

			buff := bytes.Buffer{}
			err := packet.Write(&buff, c.data, c.maxPayloadSize)
			if err == nil && c.shouldFail {
				t.Fatalf("write did not fail for case %d, but should have", i+1)
			}
			if err != nil && !c.shouldFail {
				t.Fatalf("write failed for case %d: %v", i+1, err)
			}

			// Try to read the same data off of the buffer.
			ret, _ := packet.Read(&buff, nil, 16384)
			if !bytes.Equal(c.data, ret) && err == nil {
				t.Fatalf("read data was not equal to sent data for case %d, expected %v, got %v", i+1, c.data, ret)
			}
		}()
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
		err := packet.WriteEncode(connW, data, msgpack.Codec, 16384, time.Second)
		if err != nil {
			t.Fatal(err)
		}
	}()

	var ret test
	err := packet.ReadDecode(connR, &ret, msgpack.Codec, 16384, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if ret.Msg != data.Msg {
		t.Fatalf("wrong return msg; expected '%s', got '%s'", data.Msg, ret.Msg)
	}
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
		err := packet.Write(connW, data, 16384)
		if err != nil {
			t.Fatal(err)
		}
	}()

	// Read with a timeout.
	_, err := packet.ReadTimeout(connR, nil, 16384, time.Nanosecond)
	if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
		t.Fatalf("expected timeout error, but error was '%v'", err)
	}

	// Read normally.
	ret, err := packet.ReadTimeout(connR, nil, 16384, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ret, data) {
		t.Fatalf("read data was not equal to sent data; expected '%v', got '%v'", data, ret)
	}
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
		func() {
			defer func() {
				e := recover()
				if e == nil && c.shouldPanic {
					t.Fatalf("expected a panic for case %d", i+1)
				} else if e != nil && !c.shouldPanic {
					t.Fatalf("panic for case %d", i+1)
				}
			}()

			buff := bytes.Buffer{}
			err := packet.Write(&buff, c.data, 16384)
			if err != nil {
				t.Fatalf("write case %d: %v", i, err)
			}

			// Try to read the same data off of the buffer.
			ret, err := packet.Read(&buff, c.buffer, c.maxPayloadSize)
			if err == nil && c.shouldFail {
				t.Fatalf("read did not fail for case %d, but should have", i+1)
			}
			if err != nil {
				if !c.shouldFail {
					t.Fatalf("read failed for case %d: %v", i+1, err)
				}
			} else if !bytes.Equal(c.data, ret) {
				t.Fatalf("read data was not equal to sent data for case %d, expected %v, got %v", i+1, c.data, ret)
			}
		}()
	}
}
