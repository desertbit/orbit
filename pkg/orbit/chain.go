/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2020 Sebastian Borchers <sebastian[at]desertbit.com>
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

package orbit

import (
	"sync"
)

// The chainChan type is a channel used in the chain
// that transports chainData.
type chainChan chan chainData

// The chainData type contains the Data of a Control
// Call alongside an ErrorCode that describes the outcome
// of this call.
type chainData struct {
	// The Data of a call.
	Data *Data
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
	mutex sync.Mutex
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

// new creates a new channel, adds it to the chain and returns
// the channel along with its id. The id can later be used
// to retrieve or delete the channel from the chain.
func (c *chain) new() (id uint64, cc chainChan) {
	// Create new channel.
	cc = make(chainChan)

	c.mutex.Lock()
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
	c.mutex.Unlock()
	return
}

// get returns the channel with the given id.
// Returns nil, if not found.
func (c *chain) get(id uint64) (cc chainChan) {
	c.mutex.Lock()
	cc = c.chanMap[id]
	c.mutex.Unlock()
	return
}

// delete deletes the channel with the given id from the chain.
// If the id does not exist, this is a no-op.
func (c *chain) delete(id uint64) {
	c.mutex.Lock()
	delete(c.chanMap, id)
	c.mutex.Unlock()
}
