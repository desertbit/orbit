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

package gen

import (
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestStrExplode(t *testing.T) {
	cases := []struct {
		src    string
		expect string
	}{
		{
			src:    "",
			expect: "",
		},
		{
			src:    "hello",
			expect: "hello",
		},
		{
			src:    "hEllo",
			expect: "h ello",
		},
		{
			src:    "Hello",
			expect: "hello",
		},
		{
			src:    "Hello hello",
			expect: "hello hello",
		},
		{
			src:    "Hello HellO",
			expect: "hello hell o",
		},
		{
			src:    "Hello HellOO",
			expect: "hello hell oo",
		},
		{
			src:    "Hello O",
			expect: "hello o",
		},
		{
			src:    "Hello Hell OO",
			expect: "hello hell oo",
		},
		{
			src:    "HELlo",
			expect: "he llo",
		},
		{
			src:    "sashimI",
			expect: "sashim i",
		},
	}

	for i, c := range cases {
		r.Equal(t, c.expect, strExplode(c.src), "test case %d", i)
	}
}
