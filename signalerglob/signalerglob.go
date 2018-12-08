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

/*
TODO
*/
package signalerglob

import (
	"sync"

	"github.com/desertbit/orbit/signaler"
)

type SignalerGlob struct {
	signalerMutex sync.RWMutex // TODO: Normal Mutex?
	signaler      []*signaler.Signaler
}

func New() *SignalerGlob {
	return &SignalerGlob{
		signaler: make([]*signaler.Signaler, 0),
	}
}

// TODO: Can a signaler be added twice?
func (sg *SignalerGlob) Add(s *signaler.Signaler) {
	sg.signalerMutex.Lock()
	sg.signaler = append(sg.signaler, s)
	sg.signalerMutex.Unlock()
}

func (sg *SignalerGlob) Remove(s *signaler.Signaler) {
	sg.signalerMutex.Lock()
	for i := range sg.signaler {
		if sg.signaler[i] == s {
			// Remove the element from the slice at the index.
			// We are using the safe way documented in the golang slice
			// tricks (https://github.com/golang/go/wiki/SliceTricks),
			// since we must avoid a memory leak due to pointers in a slice.
			copy(sg.signaler[i:], sg.signaler[i+1:])
			sg.signaler[len(sg.signaler)-1] = nil
			sg.signaler = sg.signaler[:len(sg.signaler)-1]
			break
		}
	}
	sg.signalerMutex.Unlock()
}

func (sg *SignalerGlob) TriggerSignaler(id string, data interface{}) (err error) {
	sg.signalerMutex.RLock()
	defer sg.signalerMutex.RUnlock()

	for _, s := range sg.signaler {
		// Trigger the signal with id at all signalers with the given data.
		// It is ignored, if one of the signalers does not contain the desired signal.
		err = s.TriggerSignal(id, data)
		if err != nil && err != signaler.ErrSignalNotFound {
			return
		}
	}

	return
}
