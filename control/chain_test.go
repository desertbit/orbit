/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018 Sebastian Borchers <sebastian.borchers[at].desertbit.com>
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
	"testing"
)

func BenchmarkChain_New(b *testing.B) {
	c := newChain()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := c.New()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChain_NewConcurrent(b *testing.B) {
	numberRoutines := 10
	c := newChain()

	wg := sync.WaitGroup{}
	wg.Add(numberRoutines)

	b.ReportAllocs()

	for n := 0; n < numberRoutines; n++ {
		go func() {
			for i := 0; i < b.N; i++ {
				_, _, err := c.New()
				if err != nil {
					b.Fatal(err)
				}
			}
			wg.Done()
		}()
		b.ResetTimer()
	}

	b.ResetTimer()
	wg.Wait()
}
