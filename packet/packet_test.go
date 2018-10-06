/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018 Sebastian Borchers <sebastian.borchers[at].desertbit.com>
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
	"github.com/desertbit/orbit/packet"
	"testing"
)

func TestWrite(t *testing.T) {
	data := []byte("This is some test data")

	cases := []struct{
		data []byte
		maxPayloadSize int
		buffer []byte
		shouldFail bool
	}{
		{
			data: data,
			maxPayloadSize: len(data),
			buffer: nil,
			shouldFail: false,
		},
		{
			data: data,
			maxPayloadSize: len(data),
			buffer: make([]byte, 1),
			shouldFail: false,
		},
		{
			data: data,
			maxPayloadSize: len(data),
			buffer: make([]byte, len(data)),
			shouldFail: false,
		},
		{
			data: data,
			maxPayloadSize: 1,
			buffer: nil,
			shouldFail: true,
		},
		{
			data: data,
			maxPayloadSize: 0,
			buffer: nil,
			shouldFail: true,
		},
		{
			data: []byte{},
			maxPayloadSize: 1,
			buffer: nil,
			shouldFail: false,
		},
		{
			data: []byte{},
			maxPayloadSize: 0,
			buffer: nil,
			shouldFail: false,
		},
	}

	for i, c := range cases {
		buff := bytes.Buffer{}
		err := packet.Write(&buff, c.data, c.maxPayloadSize)
		if err == nil && c.shouldFail {
			t.Fatalf("write did not fail for case %d, but should have", i)
		}
		if err != nil && !c.shouldFail {
			t.Fatalf("write failed for case %d: %v", i, err)
		}

		// Try to read the same data off of the buffer.
		ret, _ := packet.Read(&buff, nil, 16384)
		if !bytes.Equal(c.data, ret) && err == nil {
			t.Fatalf("read data was not equal to sent data for case %d, expected %v, got %v", i, c.data, ret)
		}
	}
}

func TestRead(t *testing.T) {
	data := []byte("This is some test data")

	cases := []struct{
		data []byte
		maxPayloadSize int
		buffer []byte
		shouldFail bool
	}{
		{
			data: data,
			maxPayloadSize: len(data),
			buffer: nil,
			shouldFail: false,
		},
		{
			data: data,
			maxPayloadSize: len(data),
			buffer: make([]byte, 1),
			shouldFail: false,
		},
		{
			data: data,
			maxPayloadSize: len(data),
			buffer: make([]byte, len(data)),
			shouldFail: false,
		},
		{
			data: data,
			maxPayloadSize: 1,
			buffer: nil,
			shouldFail: true,
		},
		{
			data: data,
			maxPayloadSize: 0,
			buffer: nil,
			shouldFail: true,
		},
		{
			data: []byte{},
			maxPayloadSize: 1,
			buffer: nil,
			shouldFail: false,
		},
		{
			data: []byte{},
			maxPayloadSize: 0,
			buffer: nil,
			shouldFail: false,
		},
	}

	for i, c := range cases {
		buff := bytes.Buffer{}
		err := packet.Write(&buff, c.data, 16384)
		if err != nil {
			t.Fatalf("write case %d: %v", i, err)
		}

		// Try to read the same data off of the buffer.
		ret, err := packet.Read(&buff, c.buffer, c.maxPayloadSize)
		if err == nil && c.shouldFail {
			t.Fatalf("read did not fail for case %d, but should have", i)
		}
		if err != nil {
			if !c.shouldFail {
				t.Fatalf("read failed for case %d: %v", i, err)
			}
		} else if !bytes.Equal(c.data, ret) {
			t.Fatalf("read data was not equal to sent data for case %d, expected %v, got %v", i, c.data, ret)
		}
	}
}
