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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey"
	"github.com/desertbit/orbit/sample/api"
	"github.com/desertbit/orbit/sample/auth"
)

const (
	actionOrbit                   = "show me the orbit"
	actionServerInfo              = "print info about the server"
	actionSubscribeToNewsletter   = "subscribe to newsletter"
	actionUnsubscribeToNewsletter = "unsubscribe from newsletter"
	actionJoinChatRoom            = "join the chat room"
	actionExit                    = "exit"
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
		action       string
		isSubscribed bool
	)
	for {
		options := []string{
			actionOrbit, actionServerInfo, actionExit, actionJoinChatRoom,
		}

		if isSubscribed {
			options = append(options, actionUnsubscribeToNewsletter)
		} else {
			options = append(options, actionSubscribeToNewsletter)
		}

		err = survey.AskOne(&survey.Select{
			Message: "Which action do you want to perform?",
			Options: options,
			Default: actionOrbit,
		}, &action, nil)
		if err != nil {
			log.Fatalln(err)
		}

		if s.IsClosed() {
			return
		}

		switch action {
		case actionOrbit:
			stream, err := s.OpenStream(api.ChannelOrbit)
			if err != nil {
				log.Fatalln(err)
			}

			go s.readStreamOrbit(stream)
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
		case actionJoinChatRoom:
			err = s.sig.SetSignalFilter(api.SignalChatIncomingMessage, api.ChatFilterData{Join: true})
			if err != nil {
				return
			}

			_ = chat(s)

			fmt.Println("\nLeaving Chat Room")
			err = s.sig.SetSignalFilter(api.SignalChatIncomingMessage, api.ChatFilterData{Join: false})
			if err != nil {
				return
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

func chat(s *Session) error {
	// Wait, until the user quits the chat by pressing Ctrl+C.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	var name string
	err := survey.AskOne(
		&survey.Input{Message: "What's your name?"},
		&name,
		survey.ComposeValidators(
			survey.MinLength(1),
			survey.MaxLength(20),
		),
	)
	if err != nil {
		return err
	}

	for {
		var text string
		err = survey.AskOne(&survey.Multiline{
			Message: "Enter your message:",
		}, &text, nil)
		if err != nil {
			return err
		}

		err = s.sig.TriggerSignal(api.SignalChatSendMessage, &api.ChatSignalData{
			Author:    name,
			Msg:       text,
			Timestamp: time.Now(),
		})
		if err != nil {
			return err
		}
	}
}
