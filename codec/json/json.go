/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers[at]desertbit.com>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

/*
Package json offers an implementation of the codec.Codec interface
for the json data format. It uses the "encoding/json" pkg to en-/decode
an entity to/from a byte slice.
*/
package json

import "encoding/json"

// Codec that encodes to and decodes from JSON.
var Codec = &jsonCodec{}

// The jsonCodec type is a private dummy struct used
// to implement the codec.Codec interface using JSON.
type jsonCodec struct{}

// Encode the value to a json byte slice.
// It uses the json.Marshal func.
func (j *jsonCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Decode the byte slice to a value.
// It uses the json.Unmarshal func.
func (j *jsonCodec) Decode(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}
