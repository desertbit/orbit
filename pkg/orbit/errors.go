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
	"errors"
)

var (
	// ErrClosed defines the error if a stream is unexpectedly closed.
	ErrClosed = errors.New("closed")

	// ErrInvalidVersion defines the error if the version of both peers do not match
	// during the version exchange.
	ErrInvalidVersion = errors.New("invalid version")
)

// An Error offers a way for handler functions of control calls to
// determine the information passed to the client, in case an error
// occurs. That way, sensitive information that may be contained in
// a standard error, can be hidden from the client.
// Instead, a Msg and a code can be sent back to give a non-sensitive
// explanation of the error and a code that is easy to check, to
// allow handling common errors.
type Error interface {
	// Embeds the standard go error interface.
	error

	// Msg returns a textual explanation of the error and should
	// NOT contain sensitive information about the application.
	Msg() string

	// Code returns an integer that can give a hint about the
	// type of error that occurred.
	Code() int
}

// Err constructs and returns a type that satisfies the control.Error interface
// from the given parameters.
func Err(err error, msg string, code int) Error {
	return errImpl{
		err:  err,
		msg:  msg,
		code: code,
	}
}

// The errImpl type is an internal error used by the control package,
// that satisfies the Error interface and allows us to throw them
// as well.
type errImpl struct {
	err  error
	msg  string
	code int
}

// Implements the control.Error interface.
func (e errImpl) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

// Implements the control.Error interface.
func (e errImpl) Msg() string {
	return e.msg
}

// Implements the control.Error interface.
func (e errImpl) Code() int {
	return e.code
}

// The ErrorCode type extends the standard go error by a simple
// integer code. It is returned in the Call- functions of this
// package and allows callers that use them to check for common
// errors via the code.
type ErrorCode struct {
	// The code is used to identify common errors.
	Code int

	err string
}

// Implements the error interface.
func (e ErrorCode) Error() string {
	return e.err
}
