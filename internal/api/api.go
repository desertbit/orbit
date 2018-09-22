/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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

//go:generate msgp
package api

const (
	Version = 1
)

type BindState int32

const (
	StateOff  BindState = 0
	StateOn   BindState = 1
	StateOnce BindState = 2
)

type InitStream struct {
	Channel string
}

type ControlCall struct {
	ID  string
	Key string
}

type ControlCallReturn struct {
	Key  string
	Msg  string
	Code int
}

type SetEvent struct {
	ID    string
	State BindState
}

type TriggerEvent struct {
	ID string
}
