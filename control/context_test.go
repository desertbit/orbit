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
	"errors"
	"github.com/desertbit/orbit/control"
	"net"
	"testing"
)

func TestContext_Control(t *testing.T) {
	const call = "callControl"

	peer1, peer2 := net.Pipe()
	ctrl1 := control.New(peer1, &control.Config{SendErrToCaller: true})
	ctrl2 := control.New(peer2, nil)
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
	}()
	ctrl1.Ready()
	ctrl2.Ready()

	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		// Check, if the Control is correctly set.
		if ctx.Control() != ctrl1 {
			err = errors.New("control was not set to the expected control")
		}
		return
	})
	ctrl2.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		// Check, if the Control is correctly set.
		if ctx.Control() != ctrl2 {
			err = errors.New("control was not set to the expected control")
		}
		return
	})

	_, err := ctrl1.Call(call, nil)
	if err != nil {
		t.Fatal("call 1", err)
	}

	_, err = ctrl2.Call(call, nil)
	if err != nil {
		t.Fatal("call 2", err)
	}
}

func TestContext_Decode(t *testing.T) {
	const call = "callDecode"

	peer1, peer2 := net.Pipe()
	ctrl1 := control.New(peer1, &control.Config{SendErrToCaller: true})
	ctrl2 := control.New(peer2, nil)
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
	}()
	ctrl1.Ready()
	ctrl2.Ready()

	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		var test string
		err = ctx.Decode(&test)
		return
	})

	_, err := ctrl2.Call(call, nil)
	if err.Error() != control.ErrNoContextData.Error() {
		t.Fatalf("expected err '%v', got '%v'", control.ErrNoContextData, err)
	}

	_, err = ctrl2.Call(call, []byte{54, 51, 50, 1, 5, 2})
	if err == nil {
		t.Fatal("expected decode to fail")
	}
}
