/*
 * ORBIT - Interlink Remote Applications
 * Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package control_test

import (
	"fmt"
	"github.com/desertbit/orbit/control"
	"net"
	"testing"
)

func TestControl_Call(t *testing.T) {
	peer1, peer2 := net.Pipe()
	ctrl1 := control.New(peer1, &control.Config{SendErrToCaller: true})
	ctrl2 := control.New(peer2, &control.Config{SendErrToCaller: true})
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
		_ = peer1.Close()
		_ = peer2.Close()
	}()

	id := "test"
	input := "args"
	output := "ret"

	ctrl1.AddFunc(id, func(ctx *control.Context) (data interface{}, err error) {
		var args string
		err = ctx.Decode(&args)
		if err != nil {
			err = fmt.Errorf("decode args 1: %v", err)
			return
		}

		if args != input {
			err = fmt.Errorf("decoded args 1: expected '%v', got '%v'", input, args)
			return
		}

		return output, nil
	})

	ctrl2.AddFunc(id, func(ctx *control.Context) (data interface{}, err error) {
		var args string
		err = ctx.Decode(&args)
		if err != nil {
			err = fmt.Errorf("decode args 2: %v", err)
			return
		}

		if args != input {
			err = fmt.Errorf("decoded args 2: expected '%v', got '%v'", input, args)
			return
		}

		return output, nil
	})

	ctrl1.Ready()
	ctrl2.Ready()

	ctx, err := ctrl1.Call(id, input)
	if err != nil {
		t.Fatalf("call: %v", err)
	}

	var ret string
	err = ctx.Decode(&ret)
	if err != nil {
		t.Fatalf("decode ret 1: %v", err)
	}

	if ret != output {
		t.Fatalf("decoded ret 1: expected '%v', got '%v'", output, ret)
	}

	ctx, err = ctrl2.Call(id, input)
	if err != nil {
		t.Fatalf("call: %v", err)
	}

	err = ctx.Decode(&ret)
	if err != nil {
		t.Fatalf("decode ret 2: %v", err)
	}

	if ret != output {
		t.Fatalf("decoded ret 2: expected '%v', got '%v'", output, ret)
	}
}
