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

package main

import (
	"fmt"
	"log"
	"net"

	"github.com/desertbit/orbit/codec/msgpack"

	"github.com/desertbit/orbit"
	"github.com/desertbit/orbit/roc"
	"github.com/desertbit/orbit/roe"
	"github.com/desertbit/orbit/sample/api"
)

type Session struct {
	*orbit.Session
}

func newSession(orbitSession *orbit.Session) (s *Session, err error) {
	s = &Session{
		Session: orbitSession,
	}

	// Always close the session on error.
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	// Log if the session closes.
	s.OnClose(func() error {
		log.Println("session closed")
		return nil
	})

	calls, events, err := s.Init(&orbit.Init{
		AcceptStreams: orbit.InitAcceptStreams{
			api.ChannelIDRaw:    handleStreamRaw,
			api.ChannelIDPacket: handleStreamPacket,
			api.ChannelIDEvent: func(stream net.Conn) error {
				evs := roe.New(stream, nil)
				l := evs.OnEvent(api.EventHello)
				data := <-l.C
				var s string
				msgpack.Codec.Decode(data.Data, &s)
				fmt.Println(s)
				return nil
			},
		},
		ROC: orbit.InitROC{
				Funcs: roc.Funcs{
					"takeAHugeDump": func(c *roc.Context) (interface{}, error) {
						fmt.Println("░░░░░░░░░░░█▀▀░░█░░░░░░")
						fmt.Println("░░░░░░▄▀▀▀▀░░░░░█▄▄░░░░")
						fmt.Println("░░░░░░█░█░░░░░░░░░░▐░░░")
						fmt.Println("░░░░░░▐▐░░░░░░░░░▄░▐░░░")
						fmt.Println("░░░░░░█░░░░░░░░▄▀▀░▐░░░")
						fmt.Println("░░░░▄▀░░░░░░░░▐░▄▄▀░░░░")
						fmt.Println("░░▄▀░░░▐░░░░░█▄▀░▐░░░░░")
						fmt.Println("░░█░░░▐░░░░░░░░▄░█░░░░░")
						fmt.Println("░░░█▄░░▀▄░░░░▄▀▐░█░░░░░")
						fmt.Println("░░░█▐▀▀▀░▀▀▀▀░░▐░█░░░░░")
						fmt.Println("░░▐█▐▄░░█░░░░░░▐░█▄▄░░░")
						fmt.Println("░░░▀▀░▄███▄░░░▐▄▄▄▀░░░░")
						return nil, nil
					},
				},
				Config: nil, // Optional. Can be removed from here...
		},
		ROE: orbit.InitROE{
			Config: nil,
		},
	})
	if err != nil {
		return
	}

	calls.Ready()
	events.Ready()

	err = events.SetEventFilter(api.EventFilter, api.FilterData{ID: "5"})
	if err != nil {
		return
	}

	events.OnEventFunc(api.EventFilter, func(ctx *roe.Context) {
		var data api.EventData
		err := ctx.Decode(&data)
		if err != nil {
			log.Printf("OnEventFunc: %v", err)
			return
		}

		log.Printf("received event with id %s and name %s", data.ID, data.Name)
	})

	return
}
