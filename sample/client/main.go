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
	actionConnectedClients = "print number of connected clients"
	actionExit = "exit"
)

func main() {
	s, err := New("127.0.0.1:9876", auth.SampleUsername, auth.SamplePassword)
	if err != nil {
		log.Fatalln(err)
	}

	var action string
	for {
		err = survey.AskOne(&survey.Select{
			Message: "Which action do you want to perform?",
			Options: []string{
				actionConnectedClients, actionExit,
			},
			Default: actionConnectedClients,
		}, &action, nil)
		if err != nil {
			log.Fatalln(err)
		}

		if s.IsClosed() {
			return
		}

		switch action {
		case actionConnectedClients:
			count, err := s.ConnectedClientsCount()
			if err != nil {
				log.Println(err)
				continue
			}

			if count == 1 {
				fmt.Printf("Oh no! You are alone right now :(")
			} else {
				fmt.Printf("Currently, %d clients are connected with you :)", count-1)
			}
		case actionExit:
			_ = s.Close()
			fmt.Println("bye!")
			return
		}
		fmt.Println()
		fmt.Println()
	}
}
