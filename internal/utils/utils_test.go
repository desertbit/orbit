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

package utils_test

import (
	"testing"

	"github.com/desertbit/orbit/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestRandomString(t *testing.T) {
	t.Parallel()

	testCases := []uint{0, 1, 10, 100}

	for i, c := range testCases {
		s, err := utils.RandomString(c)
		require.NoErrorf(t, err, "case %d", i)
		require.Lenf(t, s, int(c), "case %d", i)
	}
}

/*func TestFirstUpper(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		val string
		exp string
	}{
		{val: "", exp: ""},
	}
}*/

func TestNoTitle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		val string
		exp string
	}{
		{val: "", exp: ""}, // 0
		{val: "hello", exp: "hello"},
		{val: "Hello", exp: "hello"},
		{val: "HELLO", exp: "hELLO"},
		{val: " Hello", exp: " Hello"},
		{val: "He llo", exp: "he llo"}, // 5
	}

	for i, c := range testCases {
		require.Exactly(t, c.exp, utils.NoTitle(c.val), "case %d", i)
	}
}
