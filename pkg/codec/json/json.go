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

/*
Package json offers an implementation of the codec.Codec interface
for the json data format. It uses the https://golang.org/pkg/encoding/json/
pkg to en-/decode an entity to/from a byte slice.
*/
package json

import "encoding/json"

// Codec that encodes to and decodes from JSON.
var Codec = &jsonCodec{}

// The jsonCodec type is a private dummy struct used
// to implement the codec.Codec interface using JSON.
type jsonCodec struct{}

// Implements the codec.Codec interface.
// It uses the json.Marshal func.
func (j *jsonCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Implements the codec.Codec interface.
// It uses the json.Unmarshal func.
func (j *jsonCodec) Decode(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}
