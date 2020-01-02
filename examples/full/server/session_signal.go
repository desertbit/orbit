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
	"log"
	"time"

	"github.com/desertbit/orbit/examples/full/api"
)

func (s *Session) timeBombRoutine() {
	var (
		args = api.TimeBombData{
			Countdown: 5,
			Gift:      asciiGift,
		}
		ticker = time.NewTicker(time.Second)
		err    error
	)
	defer ticker.Stop()

	for range ticker.C {
		// Trigger the event.
		err = s.sig.TriggerSignal(api.SignalTimeBomb, &args)
		if err != nil {
			log.Printf("timeBombRoutine, triggerSignal: %v", err)
			return
		}

		if args.HasDetonated {
			// End.
			return
		}

		if args.Gift != "" {
			// Only send the gift once.
			args.Gift = ""
		}

		args.Countdown--
		args.HasDetonated = args.Countdown == 0
	}
}

// Source: https://asciiart.website/index.php?art=holiday/christmas/other
const asciiGift = `
              .__.      .==========.
            .(\\//).  .-[ for you! ]
           .(\\()//)./  '=========='
       .----(\)\/(/)----.
       |     ///\\\     |
       |    ///||\\\    |
       |   //` + "`||||`" + `\\   |
       |      ||||      |
       |      ||||      |
       |      ||||      |
       |      ||||      |
       |      ||||      |
       |      ||||      |
       '------====------'
`
