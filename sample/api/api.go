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

//go:generate msgp
package api

const (
	ChannelIDRaw    = "raw"
	ChannelIDPacket = "packet"
	ChannelIDSignal = "signal"

	SignalHello  = "hello"
	SignalTimeBomb = "timeBomb"
	SignalFilter = "filter"
)

type AuthRequest struct {
	Username string
	Pw       []byte
}

type AuthResponse struct {
	Ok bool
}

type FilterData struct {
	ID string
}

type TimeBombData struct {
	Countdown int
	DetonationForce int
	HasDetonated bool
	DetonationImage string
}

type SignalData struct {
	ID   string
	Name string
}
