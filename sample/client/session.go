/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers@desertbit.com>
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
	"sync"
	"time"

	"github.com/desertbit/orbit/events"
	"github.com/desertbit/orbit/sample/api"

	"github.com/desertbit/orbit"
)

type Session struct {
	*orbit.Session
}

func NewSession(remoteAddr string) (s *Session, err error) {
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return
	}

	orbitSession, err := orbit.ClientSession(conn, nil)
	if err != nil {
		return
	}

	s = &Session{
		Session: orbitSession,
	}

	// Always close the session on error.
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	wg := &sync.WaitGroup{}

	controls, _, err := s.Init(&orbit.Init{
		Controls: orbit.InitControls{
			"control": {
				Config: nil, // Optional. Can be removed from here...
			},
		},
	})
	if err != nil {
		return
	}

	ctrl := controls["control"]
	ctrl.Ready()

	fmt.Print("requesting a huge dump...")
	_, err = ctrl.Call("takeAHugeDump", nil)
	if err != nil {
		return
	}
	fmt.Println(" done!")
	fmt.Println()

	// TODO: Improve to real application
	eventStream, err := s.OpenStream(api.ChannelIDEvent)
	if err != nil {
		return
	}
	evs := events.New(eventStream, nil)
	evs.AddEvent(api.HelloEvent)

	go func() {
		time.Sleep(2 * time.Second)
		err = evs.TriggerEvent(api.HelloEvent, "EVENT: hello world")
		if err != nil {
			log.Println(err)
		}
	}()

	// Open a new custom stream to the peer.
	streamRaw, err := s.OpenStream(api.ChannelIDRaw)
	if err != nil {
		return
	}
	wg.Add(1)
	go streamRawRoutine(streamRaw, wg)
	// Wait for stream to close.
	wg.Wait()

	// Open a new custom stream to the peer.
	streamPacket, err := s.OpenStream(api.ChannelIDPacket)
	if err != nil {
		return
	}
	wg.Add(1)
	go streamPacketRoutine(streamPacket, wg)
	// Wait for stream to close.
	wg.Wait()

	s.Close()
	time.Sleep(time.Second)

	return
}
