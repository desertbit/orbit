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

package main

import (
	"fmt"
	"log"
	"net"

	"github.com/desertbit/orbit/codec/msgpack"

	"github.com/desertbit/orbit"
	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/sample/api"
	"github.com/desertbit/orbit/signaler"
)

type Session struct {
	*orbit.Session

	server *Server

	ctrl *control.Control
	sig *signaler.Signaler
}

func newSession(server *Server, orbitSession *orbit.Session) (s *Session, err error) {
	s = &Session{
		Session: orbitSession,
		server: server,
	}

	// Always close the session on error.
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	// Log if the session closes.
	s.OnClose(func() error {
		return nil
	})

	s.ctrl, s.sig, err = s.Init(&orbit.Init{
		AcceptStreams: orbit.InitAcceptStreams{
			api.ChannelIDRaw:    handleStreamRaw,
			api.ChannelIDPacket: handleStreamPacket,
			api.ChannelIDSignal: func(stream net.Conn) error {
				evs := signaler.New(stream, nil)
				l := evs.OnSignal(api.SignalHello)
				data := <-l.C
				var s string
				msgpack.Codec.Decode(data.Data, &s)
				fmt.Println(s)
				return nil
			},
		},
		Control: orbit.InitControl{
			Funcs: control.Funcs{
				"takeAHugeDump": func(c *control.Context) (interface{}, error) {
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
				api.ControlConnectedClientsCount: s.connectedClientsCount,
			},
			Config: nil, // Optional. Can be removed from here...
		},
		Signaler: orbit.InitSignaler{
			Config: nil,
			Signals: []orbit.InitSignal{
				{
					ID: api.SignalTimeBomb,
				},
			},
		},
	})
	if err != nil {
		return
	}

	s.ctrl.Ready()
	s.sig.Ready()

	err = s.sig.SetSignalFilter(api.SignalFilter, api.FilterData{ID: "5"})
	if err != nil {
		return
	}

	s.sig.OnSignalFunc(api.SignalFilter, func(ctx *signaler.Context) {
		var data api.SignalData
		err := ctx.Decode(&data)
		if err != nil {
			log.Printf("OnSignalFunc: %v", err)
			return
		}

		log.Printf("received signal with id %s and name %s", data.ID, data.Name)
	})

	return
}
