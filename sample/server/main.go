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

	"github.com/AlecAivazis/survey"
	"github.com/desertbit/orbit/sample/auth"
)

const (
	actionPrintClientInfo = "print info about all clients"
	actionTimeBomb        = "send a gift to all clients"
	actionSendNewsletter  = "send newsletter to subscribed clients"
	actionExit            = "exit"
)

func main() {
	// WARNING: Just a showcase!
	authHook := func(username string) (hash []byte, ok bool) {
		if auth.SampleUsername != username {
			return nil, false
		}

		return auth.SamplePasswordHash, true
	}

	s, err := NewServer("127.0.0.1:9876", authHook)
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
				actionPrintClientInfo, actionTimeBomb, actionSendNewsletter, actionExit,
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
			// TODO: replace with global events
			sessions := s.Sessions()
			for _, s := range sessions {
				go s.timeBombRoutine()
			}
			fmt.Printf("sent %d gifts to our clients...\n", len(sessions))

		case actionSendNewsletter:
			// TODO: replace with global events
			sessions := s.Sessions()
			for _, s := range sessions {
				err = s.SendNewsletter(
					"Weekly Newsletter!",
					"Dear client,\nThank you for subscribing to our weekly newsletter!\nThis weeks special news: the golang orbit package! Never has networking been easier and faster!\nCheck it out: https://github.com/desertbit/orbit/\n\nYour faithful server.",
				)
				if err != nil {
					log.Fatalln(err)
				}
			}
			fmt.Print("sent out newsletter...")
		case actionExit:
			_ = s.Close()
			fmt.Println("bye!")
			return
		}

		fmt.Println()
		fmt.Println()
	}
}
