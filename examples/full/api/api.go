/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2018 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2018 Sebastian Borchers <sebastian[at]desertbit.com>
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

import (
	"time"
)

const (
	ChannelOrbit = "Orbit"

	ControlServerInfo = "ServerInfo"
	ControlClientInfo = "ClientInfo"

	SignalTimeBomb            = "TimeBomb"
	SignalNewsletter          = "Newsletter"
	SignalChatIncomingMessage = "ChatIncomingMessage"
	SignalChatSendMessage     = "ChatSendMessage"
)

type AuthRequest struct {
	Username string
	Pw       []byte
}

type AuthResponse struct {
	Ok bool
}

type ServerInfoRet struct {
	RemoteAddr   string
	Uptime       time.Time
	ClientsCount int
}

type ClientInfoRet struct {
	RemoteAddr string
	Uptime     time.Time
}

type TimeBombData struct {
	Countdown    int
	HasDetonated bool
	Gift         string
}

type NewsletterFilterData struct {
	Subscribe bool
}

type NewsletterSignalData struct {
	Subject string
	Msg     string
}

type ChatFilterData struct {
	Join bool
}

type ChatSignalData struct {
	Author    string
	Msg       string
	Timestamp time.Time
}
