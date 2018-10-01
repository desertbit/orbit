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

/*
Package api contains types that are internally used to send data via the
control and signaler pkg.
*/
//go:generate msgp
package api

const (
	// The version of the application.
	Version = 1
)

type InitStream struct {
	Channel string
}

type ControlCall struct {
	ID  string
	Key uint64
}

type ControlReturn struct {
	Key  uint64
	Msg  string
	Code int
}

type SetSignal struct {
	ID     string
	Active bool
}

type TriggerSignal struct {
	ID   string
	Data []byte
}

type SetSignalFilter struct {
	ID   string
	Data []byte
}
