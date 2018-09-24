/*
 *  ORBIT - Interlink Remote Applications
 *  Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 *  Copyright (C) 2018  Sebastian Borchers <sebastian.borchers@desertbit.com>
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
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/desertbit/closer"
)

func TestSessions(t *testing.T) {
	var wg sync.WaitGroup

	mainCloser := closer.New()
	defer mainCloser.Close()

	errChan := make(chan error, 1)
	handleError := func(err error) {
		if err == nil {
			return
		}
		select {
		case errChan <- err:
		default:
		}
	}

	// Server
	serverRunning := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer mainCloser.Close()

		ln, err := net.Listen("tcp", "127.0.0.1:9876")
		if err != nil {
			handleError(err)
			return
		}

		l := NewServer(ln, nil)
		go func() {
			select {
			case <-mainCloser.CloseChan():
			case <-l.CloseChan():
			}
			l.Close()
		}()
		go func() {
			handleError(l.Listen())
		}()

		close(serverRunning)

		var s *Session
		select {
		case <-l.CloseChan():
			return
		case <-time.After(3 * time.Second):
			handleError(fmt.Errorf("listener new session timeout"))
			return
		case s = <-l.NewSessionChan():
		}

		// TODO:

		s.Ready()
	}()

	// Wait for server startup.
	select {
	case err := <-errChan:
		t.Error(err)
		return
	case <-serverRunning:
	}

	// Client
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer mainCloser.Close()

		conn, err := net.Dial("tcp", "127.0.0.1:9876")
		if err != nil {
			handleError(err)
			return
		}

		s, err := ClientSession(conn, nil)
		if err != nil {
			handleError(err)
			return
		}

		go func() {
			select {
			case <-mainCloser.CloseChan():
			case <-s.CloseChan():
			}
			s.Close()
		}()

		// TODO:

		s.Ready()
	}()

	wg.Wait()

	select {
	case err := <-errChan:
		t.Error(err)
		return
	default:
	}
}
