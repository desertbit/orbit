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

package orbit

import (
	"errors"
	"fmt"

	"github.com/desertbit/orbit/pkg/codec"
)

var (
	// ErrNoData defines the error if no context data is available.
	ErrNoData = errors.New("no data available to decode")
)

type Data struct {
	codec codec.Codec

	// Raw is the byte representation of the encoded data.
	Raw []byte
}

// newData creates a new Data from the given Control and the
// payload data.
func newData(data []byte, codec codec.Codec) *Data {
	return &Data{
		codec: codec,

		Raw: data,
	}
}

// Decode the context data to a custom value.
// The value has to be passed as pointer.
// Returns ErrNoContextData, if there is no context data available to decode.
func (d *Data) Decode(v interface{}) error {
	// Check if no data was passed.
	if len(d.Raw) == 0 {
		return ErrNoData
	}

	// Decode the data.
	err := d.codec.Decode(d.Raw, v)
	if err != nil {
		return fmt.Errorf("data decode: %v", err)
	}

	return nil
}
