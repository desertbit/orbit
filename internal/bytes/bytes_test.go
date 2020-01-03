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

package bytes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUInt16Conversion(t *testing.T) {
	uint16Max := uint16(1<<16 - 1)
	numbers := []uint16{5251, uint16Max, 0, 1, 101, 2387, 219}

	for _, i := range numbers {
		data := FromUint16(i)
		require.Len(t, data, 2)

		ii, err := ToUint16(data)
		require.NoError(t, err)
		require.Equal(t, i, ii)
	}

	ii, err := ToUint16(nil)
	require.Equal(t, ErrInvalidLen, err)
	require.EqualValues(t, 0, ii)

	ii, err = ToUint16(make([]byte, 1))
	require.Equal(t, ErrInvalidLen, err)
	require.EqualValues(t, 0, ii)
}

func TestUInt32Conversion(t *testing.T) {
	uint32Max := uint32(1<<32 - 1)
	numbers := []uint32{5251, uint32Max, 0, 1, 101, 2387, 219}

	for _, i := range numbers {
		data := FromUint32(i)
		require.Len(t, data, 4)

		ii, err := ToUint32(data)
		require.NoError(t, err)
		require.Equal(t, i, ii)
	}

	ii, err := ToUint32(nil)
	require.Equal(t, ErrInvalidLen, err)
	require.EqualValues(t, 0, ii)

	ii, err = ToUint32(make([]byte, 1))
	require.Equal(t, ErrInvalidLen, err)
	require.EqualValues(t, 0, ii)
}

func TestUInt64Conversion(t *testing.T) {
	uint64Max := uint64(1<<64 - 1)
	numbers := []uint64{5251, uint64Max, 0, 1, 101, 2387, 219, 213613871263}

	for _, i := range numbers {
		data := FromUint64(i)
		require.Len(t, data, 8)

		ii, err := ToUint64(data)
		require.NoError(t, err)
		require.Equal(t, i, ii)
	}

	ii, err := ToUint64(nil)
	require.Equal(t, ErrInvalidLen, err)
	require.EqualValues(t, 0, ii)

	ii, err = ToUint64(make([]byte, 1))
	require.Equal(t, ErrInvalidLen, err)
	require.EqualValues(t, 0, ii)
}
