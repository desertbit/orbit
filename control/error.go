/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package control

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
