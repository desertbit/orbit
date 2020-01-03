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
Package msgpack offers an implementation of the codec.Codec interface
for the msgpack data format.

It uses the faster https://github.com/tinylib/msgp/msgp Un-/Marshaler,
if it is implemented on the entity. Otherwise, it falls back to
using the Un-/Marshal funcs from the https://gopkg.in/vmihailenco/msgpack.v3 package.
*/
package msgpack

import (
	"github.com/tinylib/msgp/msgp"
	msgpack "gopkg.in/vmihailenco/msgpack.v3"
)

// Codec that encodes to and decodes from msgpack.
var Codec = &msgpackCodec{}

// The msgpackCodec type is a private dummy struct used
// to implement the codec.Codec interface using msgpack.
type msgpackCodec struct{}

// Implements the codec.Codec interface.
// It uses the faster msgp.Marshaler if implemented.
func (mc *msgpackCodec) Encode(v interface{}) ([]byte, error) {
	if d, ok := v.(msgp.Marshaler); ok {
		return d.MarshalMsg(nil)
	}

	return msgpack.Marshal(v)
}

// Implements the codec.Codec interface.
// It uses the faster msgp.Unmarshaler if implemented.
func (mc *msgpackCodec) Decode(b []byte, v interface{}) error {
	if d, ok := v.(msgp.Unmarshaler); ok {
		_, err := d.UnmarshalMsg(b)
		return err
	}

	return msgpack.Unmarshal(b, v)
}
