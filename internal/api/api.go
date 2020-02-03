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

const (
	// The version of the application.
	Version = 2
)

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

type StreamType int

const (
	StreamTypeRaw       = 0
	StreamTypeCallAsync = 1
	StreamTypeCallInit  = 2
)

type InitStream struct {
	Service string
	ID      string
	Type    StreamType
}

type Call struct {
	Service string
	ID      string
	Key     uint32
}

type CallReturn struct {
	Key  uint32
	Msg  string
	Code int
}

type CallCancel struct {
	Key uint32
}
