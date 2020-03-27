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
	"errors"
)

var (
	// ErrClosed defines the error if a stream, session or service is closed.
	ErrClosed = errors.New("closed")

	// ErrInvalidVersion defines the error if the version of both peers do not match
	// during the version exchange.
	ErrInvalidVersion = errors.New("invalid version")

	// ErrCatchedPanic defines the error if a panic has been catched while executing user code.
	ErrCatchedPanic = errors.New("catched panic")
)

// An Error offers a way for handler functions of rpc calls to
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

// Err constructs and returns a type that satisfies the Error interface
// from the given parameters.
func Err(err error, msg string, code int) Error {
	return errImpl{
		err:  err,
		msg:  msg,
		code: code,
	}
}

// Implements the Error interface.
type errImpl struct {
	err  error
	msg  string
	code int
}

func (e errImpl) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e errImpl) Msg() string {
	return e.msg
}

func (e errImpl) Code() int {
	return e.code
}
