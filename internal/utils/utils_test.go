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

package utils_test

import (
	"testing"

	"github.com/desertbit/orbit/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestRandomString(t *testing.T) {
	testCases := []uint{0, 1, 10, 100}

	for i, c := range testCases {
		s, err := utils.RandomString(c)
		require.NoErrorf(t, err, "case %d", i+1)
		require.Lenf(t, s, int(c), "case %d", i+1)
	}
}

func TestIsOneOfStr(t *testing.T) {
	testCases := []struct {
		s    string
		poss []string
		res  bool
	}{
		{
			s:    "",
			poss: []string{},
			res:  false,
		},
		{
			s:    "",
			poss: []string{"test"},
			res:  false,
		},
		{
			s:    "",
			poss: nil,
			res:  false,
		},
		{
			s:    "test",
			poss: []string{},
			res:  false,
		},
		{
			s:    "test",
			poss: []string{"hallo"},
			res:  false,
		},
		{
			s:    "test",
			poss: []string{"hel", "test"},
			res:  true,
		},
	}

	for i, c := range testCases {
		require.Equal(t, c.res, utils.IsOneOfStr(c.s, c.poss...), "test case %d", i)
	}
}
