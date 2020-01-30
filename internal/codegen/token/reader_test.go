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

package token_test

import (
	"io"
	"strings"
	"testing"

	"github.com/desertbit/orbit/internal/codegen/token"
	"github.com/stretchr/testify/require"
)

func TestReader_Next(t *testing.T) {
	t.Parallel()

	const data = ` this 
is some ex123ample{ text


[ to test ]{: =: } the
    tokenizer   汉字 }haha`

	cases := []struct {
		val  string
		line int
		err  error
	}{
		{val: "this", line: 1, err: nil}, // 0
		{val: "is", line: 2, err: nil},
		{val: "some", line: 2, err: nil},
		{val: "ex123ample", line: 2, err: nil},
		{val: "{", line: 2, err: nil},
		{val: "text", line: 2, err: nil}, // 5
		{val: "[", line: 5, err: nil},
		{val: "to", line: 5, err: nil},
		{val: "test", line: 5, err: nil},
		{val: "]", line: 5, err: nil},
		{val: "{", line: 5, err: nil}, // 10
		{val: ":", line: 5, err: nil},
		{val: "=", line: 5, err: nil},
		{val: ":", line: 5, err: nil},
		{val: "}", line: 5, err: nil},
		{val: "the", line: 5, err: nil}, // 15
		{val: "tokenizer", line: 6, err: nil},
		{val: "汉字", line: 6, err: nil},
		{val: "}", line: 6, err: nil},
		{val: "haha", line: 6, err: nil},
		{err: io.EOF}, // 20
	}
	tr := token.NewReader(strings.NewReader(data))

	for i, c := range cases {
		tk, err := tr.Next()
		require.Equal(t, c.err, err, "case %d", i)

		if c.err != nil {
			require.Nil(t, tk)
		} else {
			require.Equal(t, c.val, tk.Value, "case %d", i)
			require.Equal(t, c.line, tk.Line, "case %d", i)
		}
	}
}

func TestReader_Reset(t *testing.T) {
	t.Parallel()

	tr := token.NewReader(strings.NewReader("test{ 1"))
	tk, err := tr.Next()
	require.NoError(t, err)
	require.Equal(t, "test", tk.Value)
	require.Equal(t, 1, tk.Line)

	tr.Reset(strings.NewReader("\nsecond test"))
	tk, err = tr.Next()
	require.NoError(t, err)
	require.Equal(t, "second", tk.Value)
	require.Equal(t, 2, tk.Line)
}
