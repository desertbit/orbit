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

package signaler_test

import (
	"testing"
	"time"

	"github.com/desertbit/orbit/signaler"
)

func TestGroup(t *testing.T) {
	sigServer1, sigClient2 := testSignaler(nil, nil)
	sigServer3, sigClient4 := testSignaler(nil, nil)
	defer closeSignalers(sigServer1, sigClient2, sigServer3, sigClient4)

	const signal = "test"
	data := "testData"

	resChan := make(chan string, 2)
	errChan := make(chan error, 2)

	handlerF := func(ctx *signaler.Context) {
		var data string
		err := ctx.Decode(&data)
		if err != nil {
			errChan <- err
			return
		}

		resChan <- data
	}

	sigServer1.AddSignal(signal)
	sigServer3.AddSignal(signal)

	sigClient2.OnSignalFunc(signal, handlerF)
	sigClient4.OnSignalFunc(signal, handlerF)

	// IMPORTANT! Wait a short time, since the signal needs to be activated
	// with the remote peer, which needs some I/O ops.
	time.Sleep(time.Millisecond * 2)

	// Create the group.
	group := signaler.NewGroup()
	group.Add(sigServer1, sigServer3)

	// Normal test, both clients should receive the signal.
	checkErr(t, "trigger 1", group.Trigger(signal, data))

	timeout := 10 * time.Millisecond
	timer := time.NewTimer(timeout)
	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			t.Fatalf("error 1: %v", err)
		case res := <-resChan:
			assert(t, res == data, "invalid result; expected '%v', got '%v'", data, res)
		case <-timer.C:
			t.Fatal("timeout 1")
		}
		timer.Reset(timeout)
	}
	_ = timer.Stop()

	// Exclude sigServer3 from trigger, only sigClient2 should receive the signal.
	checkErr(t, "trigger 2", group.Trigger(signal, data, sigServer3))

	// Get result from sigClient2.
	select {
	case err := <-errChan:
		t.Fatalf("error 2: %v", err)
	case res := <-resChan:
		assert(t, res == data, "invalid result; expected '%v', got '%v'", data, res)
	case <-time.After(timeout):
		t.Fatal("timeout 2")
	}

	// Ensure sigClient4 does not send something back.
}
