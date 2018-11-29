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
	"log"
	"net"
	"os"
	"testing"
)

const (
	Call = "Call"
)

var (
	peer1, peer2 net.Conn
	ctrl1, ctrl2 *control.Control
)

func TestMain(m *testing.M) {
	// Setup
	peer1, peer2 = net.Pipe()
	ctrl1 = control.New(peer1, &control.Config{SendErrToCaller: true})
	ctrl2 = control.New(peer2, &control.Config{SendErrToCaller: true})

	// Run
	code := m.Run()

	// Teardown
	err := ctrl1.Close()
	if err != nil {
		log.Fatalf("close ctrl1: %v", err)
	}
	err = ctrl2.Close()
	if err != nil {
		log.Fatalf("close ctrl2: %v", err)
	}

	os.Exit(code)
}

func TestControl_Call(t *testing.T) {
	input := "args"
	output := "ret"
	var ret string

	f := func(ctx *control.Context) (data interface{}, err error) {
		var args string
		err = ctx.Decode(&args)
		if err != nil {
			err = fmt.Errorf("decode args: %v", err)
			return
		}

		if args != input {
			err = fmt.Errorf("decoded args: expected '%v', got '%v'", input, args)
			return
		}

		return output, nil
	}

	ctrl1.AddFunc(Call, f)
	ctrl2.AddFunc(Call, f)

	ctrl1.Ready()
	ctrl2.Ready()

	ctx, err := ctrl1.Call(Call, input)
	checkErr(t, "call 1: %v", err)
	checkErr(t, "decode ret 1: %v", ctx.Decode(&ret))
	assert(t, ret == output, "decoded ret 1: expected '%v', got '%v'", output, ret)

	ctx, err = ctrl2.Call(Call, input)
	checkErr(t, "call 2: %v", err)
	checkErr(t, "decode ret 2: %v", ctx.Decode(&ret))
	assert(t, ret == output, "decoded ret 2: expected '%v', got '%v'", output, ret)
}

func assert(t *testing.T, condition bool, fmt string, args ...interface{}) {
	if !condition {
		t.Fatalf(fmt, args...)
	}
}

func checkErr(t *testing.T, fmt string, err error) {
	if err != nil {
		t.Fatalf(fmt, err)
	}
}
