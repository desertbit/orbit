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

package main

import (
	"fmt"
	"log"
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
