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

// Hook allows third-party code to hook into orbit's logic, to implement for example
// logging or authentication functionality.
// Hooks that return an error have the capability to abort the action that triggered them.
// E.g. if OnNewSession() returns a non-nil error, the session will be closed.
type Hook interface {
	// Server

	// Called after the session has established a first stream and
	// performed the handshake.
	// If the returned err != nil, the new session is closed immediately.
	OnNewSession(s *Session, stream Stream) error

	// Session

	// todo:
	// If the returned err != nil, the call is aborted.
	OnCall(s *Session, service, id string) error
	// todo:
	// If err == nil, then the call completed successfully.
	OnCallCompleted(s *Session, service, id string, err error)

	// If the returned err != nil, the stream is aborted.
	OnNewStream(s *Session, service, id string) error
}
