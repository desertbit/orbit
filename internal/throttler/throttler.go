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

package throttler

import (
	"time"
)

type Throttler struct {
	duration      time.Duration
	sub           time.Duration
	lastTimestamp time.Time
}

func New(duration time.Duration) *Throttler {
	return &Throttler{
		duration: duration,
	}
}

func (t *Throttler) Throttle(now time.Time) (isThrottling bool) {
	t.sub = now.Sub(t.lastTimestamp)
	if t.sub < t.duration {
		return true
	}
	t.lastTimestamp = now
	return false
}

func (t *Throttler) ThrottleSleep(now time.Time) (isThrottling bool) {
	t.sub = now.Sub(t.lastTimestamp)
	if t.sub < t.duration {
		time.Sleep(t.duration - t.sub)
		return true
	}
	t.lastTimestamp = now
	return false
}
