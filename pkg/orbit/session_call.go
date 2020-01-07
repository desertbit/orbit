/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2020 Sebastian Borchers <sebastian[at]desertbit.com>
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

package orbit

import (
	"context"
	"fmt"
	"time"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/internal/packet"
	"github.com/desertbit/orbit/pkg/codec/msgpack"
)

const (
	// A default error message that is sent back to a caller of remote function,
	// in case the Func of the remote did not return an error that conforms to
	// our Error interface.
	// This is done to prevent sensitive information to leak to the outside that
	// is usually carried in normal errors.
	// TODO: Rename to defaultCallErrorMessage
	defaultErrorMessage = "method call failed"

	// The first byte send in a request to indicate a call to a remote function.
	typeCall byte = 0

	// The first byte send in a request to indicate the response from a remote
	// called function.
	typeCallReturn byte = 1

	// The first byte send in a request to indicate that a currently running
	// request should be canceled.
	typeCallCancel byte = 2
)

func (s *Session) RegisterCall(id string, f CallFunc) {
	s.callFuncsMx.Lock()
	s.callFuncs[id] = f
	s.callFuncsMx.Unlock()
}

func (s *Session) Call(ctx context.Context, id string, data interface{}) (d *Data, err error) {
	return s.call(ctx, s.callStream, id, data)
}

func (s *Session) CallAsync(ctx context.Context, id string, data interface{}) (d *Data, err error) {
	stream, err := s.openStream(ctx, "", api.StreamTypeCallAsync)
	if err != nil {
		return
	}

	return s.call(ctx, newMxStream(stream), id, data)
}

//###############//
//### Private ###//
//###############//

func (s *Session) call(ctx context.Context, s *mxStream, id string, data interface{}) (d *Data, err error) {
	// Create a new channel with its key. This will be used to send
	// the data over that forms the response to the call.
	key, channel := s.callRetChain.new()
	defer s.callRetChain.delete(key)

	// Write to the client.
	err = s.writeCall(ctx, cs, typeCall, &api.ControlCall{ID: id, Key: key}, data)
	if err != nil {
		return
	}

	// Wait, until the response has arrived, and return its result.
	return s.waitForResponse(ctx, cs, key, channel)
}

// TODO: finish
func (s *Session) writeCall(ctx context.Context, cs *controlStream, reqType byte, headerI, dataI interface{}) (err error) {
	var header, payload []byte

	// Marshal the header data.
	header, err = msgpack.Codec.Encode(headerI)
	if err != nil {
		return fmt.Errorf("control write encode header: %v", err)
	}

	// Marshal the payload data with the configured codec,
	// unless the payload is a byte slice. Use that directly.
	if dataI != nil {
		switch v := dataI.(type) {
		case []byte:
			payload = v

		default:
			payload, err = c.codec.Encode(dataI)
			if err != nil {
				return fmt.Errorf("control write encode payload: %v", err)
			}
		}
	}

	// Ensure only one write happens at a time on the stream.
	cs.writeMx.Lock()
	defer cs.writeMx.Unlock()

	// Set the deadline when all write operations must be finished.
	deadline, ok := ctx.Deadline()
	if ok {
		err = cs.stream.SetWriteDeadline(deadline)
		if err != nil {
			return err
		}
	}

	// Write the request type.
	_, err = cs.stream.Write([]byte{reqType})
	if err != nil {
		return err
	}

	// Write the header.
	err = packet.Write(cs.stream, header)
	if err != nil {
		return err
	}

	// Write the payload.
	err = packet.Write(cs.stream, payload)
	if err != nil {
		return err
	}

	// Reset the deadline.
	return cs.stream.SetWriteDeadline(time.Time{})
}
