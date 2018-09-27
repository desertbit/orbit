/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers[at]desertbit.com>
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

package control

import (
	"sync"
)

const (
	chainIDLength = 16
)

type chainChan chan interface{}

type chain struct {
	chanMapMutex sync.Mutex
	chanMap map[uint64]chainChan
	idCount uint64
}

func newChain() *chain {
	return &chain{
		chanMap: make(map[uint64]chainChan),
		idCount: 1, // Set to 1 to leave 0 free for special purposes.
	}
}

func (c *chain) New() (id uint64, cc chainChan, err error) {
	// Create new channel.
	cc = make(chainChan)

	c.chanMapMutex.Lock()
	// Use the current id counter as new ID.
	id = c.idCount
	// Create next ID. Increment by 2 to avoid 0.
	c.idCount += 2
	// Assign channel to map with our ID.
	c.chanMap[id] = cc
	c.chanMapMutex.Unlock()
	return
}

// Returns nil if not found.
func (c *chain) Get(id uint64) (cc chainChan) {
	c.chanMapMutex.Lock()
	cc = c.chanMap[id]
	c.chanMapMutex.Unlock()
	return
}

func (c *chain) Delete(id uint64) {
	c.chanMapMutex.Lock()
	delete(c.chanMap, id)
	c.chanMapMutex.Unlock()
}
