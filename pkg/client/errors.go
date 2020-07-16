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

import "errors"

var (
	ErrClosed         = errors.New("closed")
	ErrNoData         = errors.New("no data available")
	ErrConnect        = errors.New("connect failed")
	ErrInvalidVersion = errors.New("invalid version")
	ErrCatchedPanic   = errors.New("catched panic")
)

// The Error type extends the standard go error by a simple
// integer code. It is returned in the Call- functions of this
// package and allows callers that use them to check for common
// errors via the code.
type Error interface {
	// Embeds the standard go error interface.
	error

	// Code returns an integer that can give a hint about the
	// type of error that occurred.
	Code() int
}

// NewError returns a new error with the given message and code.
func NewError(code int, msg string) Error {
	return errImpl{
		msg:  msg,
		code: code,
	}
}

// Implements the Error interface.
type errImpl struct {
	msg  string
	code int
}

func (e errImpl) Error() string {
	return e.msg
}

func (e errImpl) Code() int {
	return e.code
}
