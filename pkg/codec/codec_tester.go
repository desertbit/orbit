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

package codec

import (
	"encoding/gob"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type test struct {
	Name string
}

// Tester is a test helper to test a Codec.
// It encodes a test struct using the given codec and decodes
// it into a second test struct afterwards.
// It then uses the reflect pkg to check if both structs have
// the exact same values.
func Tester(t *testing.T, c Codec) {
	// Struct.
	ssrc := &test{Name: "test"}
	sdst := &test{}
	encoded, err := c.Encode(ssrc)
	require.NoError(t, err)

	err = c.Decode(encoded, sdst)
	require.NoError(t, err)
	require.Exactly(t, ssrc, sdst)

	// Int.
	isrc := 5
	idst := 0
	encoded, err = c.Encode(isrc)
	require.NoError(t, err)

	err = c.Decode(encoded, &idst)
	require.NoError(t, err)
	require.Exactly(t, isrc, idst)

	// Map.
	msrc := map[string]float32{"test": 0.85}
	mdst := make(map[string]float32)
	encoded, err = c.Encode(msrc)
	require.NoError(t, err)

	err = c.Decode(encoded, &mdst)
	require.NoError(t, err)
	require.Exactly(t, msrc, mdst)

	// Slice.
	slsrc := []rune{85, 48, 68}
	sldst := make([]rune, 0)
	encoded, err = c.Encode(slsrc)
	require.NoError(t, err)

	err = c.Decode(encoded, &sldst)
	require.NoError(t, err)
	require.Exactly(t, slsrc, sldst)

	// time.Time.
	tsrc := time.Unix(584846, 471448)
	tdst := time.Time{}
	encoded, err = c.Encode(tsrc)
	require.NoError(t, err)

	err = c.Decode(encoded, &tdst)
	require.NoError(t, err)
	require.Exactly(t, tsrc, tdst)
}

func init() {
	gob.Register(&test{})
}
