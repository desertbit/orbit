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
	"context"

	"github.com/desertbit/orbit/internal/api"
	c "github.com/json-iterator/go/benchmarks"
)

func (s *Session) Call(ctx context.Context, id string, data interface{}) (data *Data, err error) {
	// Create a new channel with its key. This will be used to send
	// the data over that forms the response to the call.
	key, channel := c.callRetChain.new()
	defer c.callRetChain.delete(key)

	// Write to the client.
	err = c.write(
		typeCall,
		&api.ControlCall{
			ID:  id,
			Key: key,
		},
		data,
	)
	if err != nil {
		return
	}

	// Wait, until the response has arrived, and return its result.
	return c.waitForResponse(key, timeout, channel, cancelChan)
}

func (s *Session) CallAsync(ctx context.Context)
