/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers[at]desertbit.com>
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

package roc

type Error interface {
	error
	Msg() string
	Code() int
}

func Err(err error, msg string, code int) Error {
	return errImpl{
		err:  err,
		msg:  msg,
		code: code,
	}
}

type errImpl struct {
	err  error
	msg  string
	code int
}

func (e errImpl) Error() string {
	return e.err.Error()
}

func (e errImpl) Msg() string {
	return e.msg
}

func (e errImpl) Code() int {
	return e.code
}

type ErrorCode struct {
	Code int

	err string
}

func (e ErrorCode) Error() string {
	return e.err
}
