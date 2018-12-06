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
	"testing"

	"github.com/desertbit/orbit/control"
)

func TestContext_Control(t *testing.T) {
	t.Parallel()

	ctrl1, ctrl2 := testControls(&control.Config{SendErrToCaller: true}, nil)
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
	}()

	const call = "callControl"
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
	checkErr(t, "call 1: %v", err)

	_, err = ctrl2.Call(call, nil)
	checkErr(t, "call 2: %v", err)
}

func TestContext_Decode(t *testing.T) {
	t.Parallel()

	ctrl1, ctrl2 := testControls(&control.Config{SendErrToCaller: true}, nil)
	defer func() {
		_ = ctrl1.Close()
		_ = ctrl2.Close()
	}()

	const call = "callDecode"
	ctrl1.AddFunc(call, func(ctx *control.Context) (data interface{}, err error) {
		var test string
		err = ctx.Decode(&test)
		return
	})

	_, err := ctrl2.Call(call, nil)
	assert(t, err.Error() == control.ErrNoContextData.Error(), "expected err '%v', got '%v'", control.ErrNoContextData, err)

	_, err = ctrl2.Call(call, []byte{54, 51, 50, 1, 5, 2})
	assert(t, err != nil, "expected decode to fail")
}
