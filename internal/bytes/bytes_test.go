/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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

package bytes

import "testing"

func TestUInt16Conversion(t *testing.T) {
	uint16Max := uint16(1<<16 - 1)
	numbers := []uint16{5251, uint16Max, 0, 1, 101, 2387, 219}

	for _, i := range numbers {
		data := FromUint16(i)
		if len(data) != 2 {
			t.Fatal()
		}

		ii, err := ToUint16(data)
		if err != nil {
			t.Fatal()
		} else if ii != i {
			t.Fatal()
		}
	}

	ii, err := ToUint16(nil)
	if err != ErrInvalidLen {
		t.Fatal()
	} else if ii != 0 {
		t.Fatal()
	}

	ii, err = ToUint16(make([]byte, 1))
	if err != ErrInvalidLen {
		t.Fatal()
	} else if ii != 0 {
		t.Fatal()
	}
}

func TestUInt32Conversion(t *testing.T) {
	uint32Max := uint32(1<<32 - 1)
	numbers := []uint32{5251, uint32Max, 0, 1, 101, 2387, 219}

	for _, i := range numbers {
		data := FromUint32(i)
		if len(data) != 4 {
			t.Fatal()
		}

		ii, err := ToUint32(data)
		if err != nil {
			t.Fatal()
		} else if ii != i {
			t.Fatal()
		}
	}

	ii, err := ToUint32(nil)
	if err != ErrInvalidLen {
		t.Fatal()
	} else if ii != 0 {
		t.Fatal()
	}

	ii, err = ToUint32(make([]byte, 1))
	if err != ErrInvalidLen {
		t.Fatal()
	} else if ii != 0 {
		t.Fatal()
	}
}

func TestUInt64Conversion(t *testing.T) {
	uint64Max := uint64(1<<64 - 1)
	numbers := []uint64{5251, uint64Max, 0, 1, 101, 2387, 219, 213613871263}

	for _, i := range numbers {
		data := FromUint64(i)
		if len(data) != 8 {
			t.Fatal()
		}

		ii, err := ToUint64(data)
		if err != nil {
			t.Fatal()
		} else if ii != i {
			t.Fatal()
		}
	}

	ii, err := ToUint64(nil)
	if err != ErrInvalidLen {
		t.Fatal()
	} else if ii != 0 {
		t.Fatal()
	}

	ii, err = ToUint64(make([]byte, 1))
	if err != ErrInvalidLen {
		t.Fatal()
	} else if ii != 0 {
		t.Fatal()
	}
}
