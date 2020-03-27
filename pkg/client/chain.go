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

package client

import (
	"sync"
)

type chainChan chan chainData

type chainData struct {
	Data []byte
	Err  error
}

type chain struct {
	mutex   sync.Mutex
	chanMap map[uint32]chainChan
	key     uint32
}

func newChain() *chain {
	return &chain{
		chanMap: make(map[uint32]chainChan),
	}
}

func (c *chain) New() (key uint32, cc chainChan) {
	cc = make(chainChan, 1)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.incrementKey()

	key = c.key
	c.chanMap[key] = cc
	return
}

func (c *chain) NewKey() (key uint32) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.incrementKey()
	return c.key
}

// Get returns the channel with the given key.
// Returns nil, if not found.
func (c *chain) Get(key uint32) (cc chainChan) {
	c.mutex.Lock()
	cc = c.chanMap[key]
	c.mutex.Unlock()
	return
}

// Delete the channel with the given key from the chain.
// If the key does not exist, this is a no-op.
func (c *chain) Delete(key uint32) {
	c.mutex.Lock()
	delete(c.chanMap, key)
	c.mutex.Unlock()
}

func (c *chain) incrementKey() {
	// Create next key.
	c.key++

	// We avoid a key of 0, as this is a special value.
	if c.key == 0 {
		c.key++
	}
}
