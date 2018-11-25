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
	"github.com/AlecAivazis/survey"
	"github.com/desertbit/orbit/sample/auth"
	"log"
)

const (
	actionPrintClientInfo = "print info about all clients"
	actionTimeBomb = "send a gift to all clients"
	actionExit = "exit"
)

func main() {
	// WARNING: Just a showcase!
	authHook := func(username string) (hash []byte, ok bool) {
		if auth.SampleUsername != username {
			return nil, false
		}

		return auth.SamplePasswordHash, true
	}

	s, err := NewServer("127.0.0.1:9876",  authHook)
	if err != nil {
		log.Fatalln(err)
	}

	// Start listening for incoming connections.
	go func() {
		err = s.Listen()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	var action string
	for {
		err = survey.AskOne(&survey.Select{
			Message: "Which action do you want to perform?",
			Options: []string{
				actionPrintClientInfo, actionTimeBomb, actionExit,
			},
			Default: actionPrintClientInfo,
		}, &action, nil)
		if err != nil {
			log.Fatalln(err)
		}

		switch action {
		case actionPrintClientInfo:
			sessions := s.Sessions()

			for _, session := range sessions {
				info, err := session.ClientInfo()
				if err != nil {
					log.Fatalln(err)
				}

				fmt.Printf("client at %v is up since %v\n", info.RemoteAddr, info.Uptime.String())
			}
		case actionTimeBomb:
			sessions := s.Sessions()
			for _, s := range sessions {
				go s.timeBombRoutine()
			}
			fmt.Printf("sent %d gifts to our clients...\n", len(sessions))
		case actionExit:
			_ = s.Close()
			fmt.Println("bye!")
			return
		}

		fmt.Println()
		fmt.Println()
	}
}
