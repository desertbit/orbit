/*
 * ORBIT - Interlink Remote Applications
 * Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"log"

	"github.com/desertbit/orbit/sample/api"
	"github.com/desertbit/orbit/signaler"
)

func (s *Session) chatFilter(ctx *signaler.Context) (f signaler.Filter, err error) {
	var fData api.ChatFilterData
	err = ctx.Decode(&fData)
	if err != nil {
		return
	}

	f = func(data interface{}) (conforms bool, err error) {
		return fData.Join, nil
	}
	return
}

func (s *Session) onChatSendMessage(ctx *signaler.Context) {
	var data api.ChatSignalData
	err := ctx.Decode(&data)
	if err != nil {
		log.Printf("onChatSendMessage: %v", err)
		return
	}

	// Distribute the signal to all chat participants.
	err = s.server.chatSigGroup.Trigger(api.SignalChatIncomingMessage, data, s.sig)
	if err != nil {
		log.Printf("onChatSendMessage: %v", err)
		return
	}
}
