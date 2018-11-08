/*
 * ORBIT - Interlink Remote Applications
 * Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (C) 2018  Sebastian Borchers <sebastian.borchers[at]desertbit.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"github.com/desertbit/orbit/sample/api"
	"github.com/desertbit/orbit/signaler"
	"log"
)

func (s *Session) onEventTimeBomb(ctx *signaler.Context) {
	var args api.TimeBombData
	err := ctx.Decode(&args)
	if err != nil {
		log.Printf("onEventTimeBomb error: %v", err)
		return
	}

	if args.HasDetonated {
		fmt.Printf("Time bomb has detonated!\nDetonation Force was: %d\nImage of the destructed site: %s\n", args.DetonationForce, args.DetonationImage)
		return
	}

	fmt.Printf("Time bomb is ticking! Countdouwn %d...\n", args.Countdown)
}
