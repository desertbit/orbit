/*
 *  Juno
 *  Copyright (C) 2016 DesertBit
 */

package bcrypt

import "testing"

func TestBCrypt(t *testing.T) {
	pw := []byte("p93rp9fbp9efb")
	hash, err := Generate(pw)
	if err != nil {
		t.Fatal(err)
	}

	err = Compare(hash, pw)
	if err != nil {
		t.Fatal(err)
	}

	pw[0] = 0
	err = Compare(hash, pw)
	if err == nil {
		t.Fatal(err)
	}

	err = Compare(hash, []byte{})
	if err == nil {
		t.Fatal()
	}

	err = Compare([]byte{}, []byte{})
	if err == nil {
		t.Fatal()
	}
}
