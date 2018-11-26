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

import (
	"time"
)

const (
	ChannelIDRaw    = "Raw"
	ChannelIDPacket = "Packet"
	ChannelIDSignal = "Signal"

	ControlServerInfo = "ServerInfo"
	ControlClientInfo = "ClientInfo"

	SignalHello      = "Hello"
	SignalTimeBomb   = "TimeBomb"
	SignalNewsletter = "Newsletter"
)

type AuthRequest struct {
	Username string
	Pw       []byte
}

type AuthResponse struct {
	Ok bool
}

type ServerInfoRet struct {
	RemoteAddr string
	Uptime time.Time
	ClientsCount int
}

type ClientInfoRet struct {
	RemoteAddr string
	Uptime time.Time
}

type TimeBombData struct {
	Countdown       int
	HasDetonated    bool
	Gift            string
}

type NewsletterFilterData struct {
	Subscribe bool
}

type NewsletterSignalData struct {
	Subject string
	Msg string
}
