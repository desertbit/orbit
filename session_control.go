/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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

package orbit

import (
	"time"
)

const (
	newControlStreamTimeout = 12 * time.Second
)

/*
// TODO: Finish this.
// Control registers a new control. This method blocks until
// the control is connected and ready.
// Returns ErrTimeout on timeout.
func (s *Session) NewControl(
	channel string,
	funcs control.Funcs,
	config *control.Config,
) {
	go func() {
		err := s.openControl(channel, funcs, config)
		if err != nil {
			// TODO:
			return
		}
	}()
}

func (s *Session) Control(channel string) (c *Control) {
	s.controlsMutex.Lock()
	c = s.controls[channel]
	s.controlsMutex.Unlock()
	return
}

func (s *Session) openControl(channel string,
	funcs control.Funcs,
	config *control.Config,
) (err error) {
	// Connect / wait for the stream connection.
	var stream net.Conn
	if s.isClient {
		stream, err = s.OpenStreamTimeout(channel, newControlStreamTimeout)
		if err != nil {
			return
		}
	} else {
		closeChan := make(chan struct{})
		defer close(closeChan)

		streamChan := make(chan net.Conn)
		s.OnNewStream(channel, func(conn net.Conn) error {
			select {
			case <-closeChan:
				conn.Close()
				return fmt.Errorf("not waiting for stream: accept disabled")
			case streamChan <- conn:
			}
			return nil
		})

		timeoutTimer := time.NewTimer(newControlStreamTimeout)
		defer timeoutTimer.Stop()

		select {
		case <-timeoutTimer.C:
			return nil, ErrTimeout
		case stream = <-streamChan:
		}
	}

	// Create the control.
	ctrl = control.New(stream, config)

	// Close the control if the session closes.
	go func() {
		select {
		case <-s.CloseChan():
		case <-ctrl.CloseChan():
		}
		ctrl.Close()
	}()

	return
}
*/
