/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018 Sebastian Borchers <sebastian.borchers[at].desertbit.com>
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

package codec

import (
	"encoding/gob"
	"reflect"
	"testing"
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
	val := &test{Name: "test"}
	to := &test{}

	encoded, err := c.Encode(val)
	if err != nil {
		t.Fatal("Encode error:", err)
	}
	err = c.Decode(encoded, to)
	if err != nil {
		t.Fatal("Decode error:", err)
	}
	if !reflect.DeepEqual(val, to) {
		t.Fatalf("Roundtrip codec mismatch, expected\n%#v\ngot\n%#v", val, to)
	}
}

func init() {
	gob.Register(&test{})
}
