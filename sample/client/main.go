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
	"github.com/desertbit/orbit/sample/api"
	"github.com/desertbit/orbit/sample/auth"
	"log"
)

const (
	actionServerInfo = "print info about the server"
	actionSubscribeToNewsletter = "subscribe to newsletter"
	actionUnsubscribeToNewsletter = "unsubscribe from newsletter"
	actionExit       = "exit"
)

func main() {
	s, err := New("127.0.0.1:9876", auth.SampleUsername, auth.SamplePassword)
	if err != nil {
		log.Fatalln(err)
	}

	// Per default, we do not want to subscribe to the newsletter
	err = s.sig.SetSignalFilter(api.SignalNewsletter, api.NewsletterFilterData{
		Subscribe: false,
	})
	if err != nil {
		return
	}

	var (
		action string
		isSubscribed bool
		options []string
	)
	for {
		if isSubscribed {
			options = []string{
				actionServerInfo, actionUnsubscribeToNewsletter, actionExit,
			}
		} else {
			options = []string{
				actionServerInfo, actionSubscribeToNewsletter, actionExit,
			}
		}

		err = survey.AskOne(&survey.Select{
			Message: "Which action do you want to perform?",
			Options: options,
			Default: actionServerInfo,
		}, &action, nil)
		if err != nil {
			log.Fatalln(err)
		}

		if s.IsClosed() {
			return
		}

		switch action {
		case actionServerInfo:
			info, err := s.ServerInfo()
			if err != nil {
				log.Println(err)
				continue
			}

			fmt.Printf("Server is at: %v\nUp since: %v\nClients connected: %d", info.RemoteAddr, info.Uptime.String(), info.ClientsCount)
		case actionSubscribeToNewsletter:
			fallthrough
		case actionUnsubscribeToNewsletter:
			isSubscribed = !isSubscribed
			err = s.sig.SetSignalFilter(api.SignalNewsletter, api.NewsletterFilterData{
				Subscribe: isSubscribed,
			})
			if err != nil {
				return
			}
			if isSubscribed {
				fmt.Print("subscribed to newsletter!")
			} else {
				fmt.Print("unsubscribed from newsletter!")
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
