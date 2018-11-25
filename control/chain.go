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

package control

import (
	"sync"
)

// The chainChan type is a channel used in the chain
// that transports chainData.
type chainChan chan chainData

// The chainData type contains the Context of a Control
// Call alongside an ErrorCode that describes the outcome
// of this call.
type chainData struct {
	// The Context of a call.
	Context *Context
	// Contains the error message and an error code
	// that the call may have produced.
	Err error
}

// The chain type contains a map of chainChan channels
// and methods to create, get and delete such channels.
// Typically, one Control uses one chain and creates
// with it one channel per request/response pair.
//
// The channels are used to return the response data
// to the calling function of the requester that
// waits until data arrives on the channel.
type chain struct {
	// Synchronises the access to the channel map and the idCount.
	chainMutex sync.Mutex
	// Stores the channels that are handled by this chain.
	chanMap map[uint64]chainChan
	// A simple counter to create an unique key for new
	// channels that is used to store them in the map.
	idCount uint64
}

// newChain creates a new chain.
func newChain() *chain {
	return &chain{
		chanMap: make(map[uint64]chainChan),
	}
}

// New creates a new channel, adds it to the chain and returns
// the channel along with its id. The id can later be used
// to retrieve or delete the channel from the chain.
func (c *chain) New() (id uint64, cc chainChan) {
	// Create new channel.
	cc = make(chainChan)

	c.chainMutex.Lock()
	// Create next ID.
	c.idCount++
	// We avoid an id of 0, as we need this special value
	// to indicate the absence of a channel in the control
	// implementation.
	if c.idCount == 0 {
		c.idCount++
	}
	// Use the current id counter as new ID.
	id = c.idCount
	// Assign the channel to the map with our ID.
	c.chanMap[id] = cc
	c.chainMutex.Unlock()
	return
}

// Get returns the channel with the given id.
// Returns nil, if not found.
func (c *chain) Get(id uint64) (cc chainChan) {
	c.chainMutex.Lock()
	cc = c.chanMap[id]
	c.chainMutex.Unlock()
	return
}

// Delete deletes the channel with the given id from the chain.
// If the id does not exist, this is a no-op.
func (c *chain) Delete(id uint64) {
	c.chainMutex.Lock()
	delete(c.chanMap, id)
	c.chainMutex.Unlock()
}
