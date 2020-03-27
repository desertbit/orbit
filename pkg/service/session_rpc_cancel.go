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

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/desertbit/orbit/internal/api"

	"github.com/desertbit/orbit/internal/rpc"
	"github.com/desertbit/orbit/pkg/transport"
)

const (
	earlyCancelLifetime = 10 * time.Second
)

func (s *session) handleCancelStream(stream transport.Stream) (err error) {
	// Close stream on error.
	defer func() {
		if err != nil {
			stream.Close()
		}
	}()

	// Abort if a cancel stream is already present.
	var hasStream bool
	s.cancelMx.Lock()
	hasStream = s.hasCancelStream
	s.hasCancelStream = true
	s.cancelMx.Unlock()
	if hasStream {
		return errors.New("a cancel stream is already present")
	}

	// Start the read routine.
	go s.rpcCancelReadRoutine(stream)
	return nil
}

func (s *session) rpcCancelReadRoutine(stream transport.Stream) {
	// Close the session on exit.
	// Currently only one cancel stream is supported.
	defer s.Close_()

	var (
		err     error
		reqType api.RPCType
		header  []byte
	)

	for {
		// Read and reuse the header buffer.
		reqType, header, _, err = rpc.Read(stream, header, nil)
		if err != nil {
			// Log errors, but only, if the session or stream are not closing.
			if !s.IsClosing() && !stream.IsClosed() && !s.conn.IsClosedError(err) {
				s.log.Error().
					Err(err).
					Msg("rpc: cancel read routine")
			}
			return
		}

		err = s.handleCancel(reqType, header)
		if err != nil {
			s.log.Warn().
				Err(err).
				Msg("rpc: failed to handle cancel request")
			continue
		}
	}
}

func (s *session) handleCancel(reqType api.RPCType, header []byte) (err error) {
	// Ensure the type is valid.
	if reqType != api.RPCTypeCancel {
		return fmt.Errorf("invalid request type: %v: expected cancel type", reqType)
	}

	// Decode the request header.
	var h api.RPCCancel
	err = api.Codec.Decode(header, &h)
	if err != nil {
		return fmt.Errorf("decode header: %w", err)
	}

	var (
		ok        bool
		cancel    context.CancelFunc
		closeChan = make(chan struct{})
	)

	// Obtain the cancel function if present.
	// If not present, then this cancel request might have arrived before the actual call request.
	// Notify the call to do an early cancel.
	s.cancelMx.Lock()
	cancel, ok = s.cancelCalls[h.Key]
	if !ok {
		s.cancelCalls[h.Key] = func() {
			close(closeChan)
		}
	}
	s.cancelMx.Unlock()

	// Cancel the call if possible.
	if ok {
		cancel()
	} else {
		// Tidy up the early cancel value from the map after a timeout.
		go func() {
			t := time.NewTimer(earlyCancelLifetime)
			defer t.Stop()

			select {
			case <-closeChan:
				return
			case <-t.C:
				s.deleteCancelFunc(h.Key)
			}
		}()
	}

	return
}

func (s *session) setCancelFunc(key uint32, f context.CancelFunc) (alreadyCanceled bool) {
	var ff context.CancelFunc

	s.cancelMx.Lock()
	ff, alreadyCanceled = s.cancelCalls[key]
	if alreadyCanceled {
		delete(s.cancelCalls, key)
	} else {
		s.cancelCalls[key] = f
	}
	s.cancelMx.Unlock()

	// Call the function which was present in the map.
	if alreadyCanceled {
		ff()
	}
	return
}

func (s *session) deleteCancelFunc(key uint32) {
	s.cancelMx.Lock()
	delete(s.cancelCalls, key)
	s.cancelMx.Unlock()
}
