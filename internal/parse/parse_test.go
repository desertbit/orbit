/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2019 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2019 Sebastian Borchers <sebastian[at]desertbit.com>
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

package parse_test

import (
	"testing"

	"github.com/desertbit/orbit/internal/parse"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	data := `
service bencher {
    call test(Plate) (stream Rect)
    revcall test2({
        i int
        v float64
        c map[int][]Rect
    }) ({
        lol string
    })
    stream hello
}

type Plate {
    version int
    name string
    rect Rect
    test map[int]Rect
    test2 []Rect
    test3 []float32
    test4 map[string]map[int][]Rect
}

type Rect {
    x1 float32
    y1 float32
    x2 float32
    y2 float32
    c  Char
}

type Char {
    lol string
}`

	_, _, err := parse.Parse(data)
	require.NoError(t, err)
}