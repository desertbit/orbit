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

package client

import (
	"errors"
	"fmt"

	"github.com/desertbit/orbit/internal/api"
	"github.com/desertbit/orbit/internal/rpc"
)

const (
	numRPCReadRoutines = 3
)

func (s *session) startRPCRoutines() {
	for i := 0; i < numRPCReadRoutines; i++ {
		go s.rpcReadRoutine()
	}
}

func (s *session) rpcReadRoutine() {
	// Close the session on exit.
	defer s.Close_()

	var (
		err     error
		reqType api.RPCType
		header  []byte
		payload []byte
	)

	for {
		s.streamReadMx.Lock()
		reqType, header, payload, err = rpc.Read(s.stream, nil, nil, s.maxHeaderSize, s.maxRetSize)
		s.streamReadMx.Unlock()
		if err != nil {
			// Log errors, but only, if the session or stream are not closing.
			if !s.IsClosing() && !s.stream.IsClosed() && !s.conn.IsClosedError(err) {
				s.log.Error().
					Err(err).
					Msg("rpc: read routine")
			}
			return
		}

		s.handleRPCRequest(reqType, header, payload)
	}
}

func (s *session) handleRPCRequest(reqType api.RPCType, headerData, payloadData []byte) {
	var err error

	// Check the request type.
	// The client supports only a limited range of requests.
	switch reqType {
	case api.RPCTypeReturn:
		err = s.handleRPCReturn(headerData, payloadData)
	default:
		err = fmt.Errorf("invalid request type '%v'", reqType)
	}
	if err != nil {
		s.log.Error().
			Err(err).
			Msg("rpc: failed to handle request")
	}
}

// handleReturn processes an incoming response with request type 'typeCallReturn'.
// It decodes the header and uses it to retrieve the correct channel for this callReturn.
// The response payload is then wrapped in a context and sent over the channel
// back to the calling function that waits for it.
// It can then deliver the response to the caller, which is the end of
// one request-response cycle.
func (s *session) handleRPCReturn(headerData, payloadData []byte) error {
	// Decode the header.
	var header api.RPCReturn
	err := api.Codec.Decode(headerData, &header)
	if err != nil {
		return fmt.Errorf("return request: decode header: %w", err)
	}

	// Get the channel by the key.
	channel := s.chain.Get(header.Key)
	if channel == nil {
		return fmt.Errorf("return request: no handler func available for key '%d'", header.Key)
	}

	// Create the channel data.
	rData := chainData{Data: payloadData}

	// Create an ErrorCode, if an error is present.
	if header.Err != "" {
		rData.Err = errImpl{msg: header.Err, code: header.ErrCode}
	}

	// Send the return data to the channel.
	select {
	case channel <- rData:
		return nil
	default:
		return errors.New("return request: failed to deliver return data: channel full")
	}
}
