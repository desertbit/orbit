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

//go:generate msgp
package api

import "github.com/desertbit/orbit/pkg/codec/msgpack"

const (
	// The version of the application.
	Version = 3
)

var (
	Codec = msgpack.Codec
)

//#################//
//### Handshake ###//
//#################//

type HandshakeCode int

const (
	HSOk             HandshakeCode = 0
	HSInvalidVersion HandshakeCode = 1
)

type HandshakeArgs struct {
	Version byte
}

type HandshakeRet struct {
	Code      HandshakeCode
	SessionID string
}

//##############//
//### Stream ###//
//##############//

type StreamType int

const (
	StreamTypeRaw         = 0
	StreamTypeAsyncCall   = 1
	StreamTypeCancelCalls = 2
)

type InitStream struct {
	ID   string
	Type StreamType
	Data map[string][]byte
}

//###########//
//### RPC ###//
//###########//

type RPCType byte

const (
	RPCTypeCall   RPCType = 0
	RPCTypeReturn RPCType = 1
	RPCTypeCancel RPCType = 2
)

type RPCCall struct {
	ID   string
	Key  uint32
	Data map[string][]byte
}

type RPCReturn struct {
	Key     uint32
	Err     string
	ErrCode int
}

type RPCCancel struct {
	Key uint32
}
