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

package token

import (
	"errors"
	"io"
	"unicode"
)

var (
	ErrNewlineInSingQuoteString = errors.New("newline in single quote string")
)

type Reader interface {
	// Returns io.EOF, if no more tokens available.
	// This especially means, that if a non nil token is returned, err is nil.
	// Other errors may be returned at any time.
	Next() (t *Token, err error)

	// Resets the reader to read from the given rune reader.
	Reset(rr io.RuneReader)
}

// Implements the Reader interface.
type reader struct {
	rr io.RuneReader

	nextValue      []rune
	nextTokenReady bool
	singQuote      bool
	lines          int
}

func NewReader(rr io.RuneReader) Reader {
	return &reader{
		rr:        rr,
		nextValue: make([]rune, 0),
		lines:     1,
	}
}

func (r *reader) Next() (t *Token, err error) {
	var nr rune

	for {
		// If a token from a previous call is available, return it first.
		if r.nextTokenReady {
			t = r.buildToken()
			r.nextTokenReady = false
			return
		}

		// Read next rune.
		nr, _, err = r.rr.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) && len(r.nextValue) > 0 {
				// End of rune reader, return last token.
				t = r.buildToken()
				// Ignore error for now.
				err = nil
			}
			return
		}

		if r.singQuote && nr != singQuote {
			// Newlines in single quotes are not allowed!
			if nr == newLine {
				err = ErrNewlineInSingQuoteString
				return
			}

			r.nextValue = append(r.nextValue, nr)
		} else if unicode.IsSpace(nr) {
			// Build the token before checking for a new line.
			if len(r.nextValue) > 0 {
				t = r.buildToken()
			}

			// Increment new line.
			if nr == newLine {
				r.lines++
			}

			if t != nil {
				return
			}
		} else if nr == braceL || nr == braceR ||
			nr == bracketL || nr == bracketR ||
			nr == colon ||
			nr == equal ||
			nr == singQuote {
			if nr == singQuote {
				r.singQuote = !r.singQuote
			}

			// Return next token first.
			if len(r.nextValue) > 0 {
				t = r.buildToken()
				// Save the rune for the next call.
				r.nextValue = append(r.nextValue, nr)
				r.nextTokenReady = true
				return
			}

			// Return this token directly.
			t = &Token{Value: string(nr), Line: r.lines}
			return
		} else {
			r.nextValue = append(r.nextValue, nr)
		}
	}
}

func (r *reader) Reset(rr io.RuneReader) {
	r.rr = rr
	r.nextValue = r.nextValue[:0]
	r.nextTokenReady = false
	r.singQuote = false
	r.lines = 1
}

func (r *reader) buildToken() (t *Token) {
	t = &Token{Value: string(r.nextValue), Line: r.lines}
	// Reset slice, but remain cap.
	r.nextValue = r.nextValue[:0]
	return
}
