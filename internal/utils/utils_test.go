/*
 * ORBIT - Interlink Remote Applications
 * Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
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

package utils_test

import (
	"testing"

	"github.com/desertbit/orbit/internal/utils"
)

func TestRandomString(t *testing.T) {
	testCases := []uint{0, 1, 10, 100}
	for i, c := range testCases {
		s, err := utils.RandomString(c)
		if err != nil {
			t.Fatalf("case %d: %v", i+1, err)
		}
		if uint(len(s)) != c {
			t.Fatalf("wrong length; expected %d, got %d", c, len(s))
		}
	}
}
