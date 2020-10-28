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

// Package lexer is based on Rop Pike's talk: https://www.youtube.com/watch?v=HxaD_trXwRE
package lexer

import (
	"fmt"
	"unicode/utf8"

	"github.com/desertbit/closer/v3"
)

// The Lexer interface represents a lexer that lexes its input
// into Tokens.
type Lexer interface {
	closer.Closer

	// Next returns the next available Token from the lexer.
	// If no more Tokens are available or the lexer was closed,
	// false is returned.
	Next() (Token, bool)
}

// stateFn represents the state of the scanner
// as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner and implements the
// Lexer interface.
type lexer struct {
	closer.Closer

	input       string // the string being scanned.
	start       int    // start position of this token.
	startLine   int    // start line of this token.
	startCol    int    // start column of this token.
	pos         int    // current position in the input.
	col         int    // current column in the input.
	width       int    // width of last rune read from input.
	line        int    // current line in the input.
	prevLineCol int    // stores col pos of previous line.
	curRune     rune   // caches the last rune retrieved by l.next.

	tokens chan Token // channel of scanned tokens.
}

// Lex concurrently starts lexing the given input
// and returns the associated Lexer instance.
func Lex(cl closer.Closer, input string) Lexer {
	l := &lexer{
		Closer:    cl,
		input:     input,
		startLine: 1,
		startCol:  1,
		col:       1,
		line:      1,
		tokens:    make(chan Token, 2),
	}

	// Concurrently start lexing.
	cl.CloserAddWait(1)
	go l.run()

	return l
}

// Implements the Lexer interface.
func (l *lexer) Next() (t Token, ok bool) {
	// Do not listen on the closing chan. This can
	// cause tokens in the channel to not be consumed.
	t, ok = <-l.tokens
	return
}

// run lexes the input by executing state functions until
// the state is nil.
// Closes the l.tokens channel, once done.
func (l *lexer) run() {
	defer l.CloseAndDone_()

	for state := lexTokenStart(l); state != nil && !l.IsClosing(); {
		state = state(l)
	}

	// No more text left, all tokens emitted.
	close(l.tokens)
}

// emit publishes a token to the client.
func (l *lexer) emit(t TokenType) {
	l.tokens <- Token{
		Type:  t,
		Value: l.input[l.start:l.pos],
		Pos:   l.curPos(),
	}
	l.start = l.pos
	l.startLine = l.line
	l.startCol = l.col
}

// next returns the next rune in the input
// or eof, if none is available.
func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	l.curRune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	if l.curRune == newline {
		l.line++
		l.prevLineCol = l.col
		l.col = 0
	}
	l.col++
	return l.curRune
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
	l.startLine = l.line
	l.startCol = l.col
	l.curRune = eof
}

// backup steps back one rune.
// Can be called only once per call of l.next.
func (l *lexer) backup() {
	if l.curRune == newline {
		l.col = l.prevLineCol
		l.prevLineCol = -1
		l.line--
	} else {
		l.col--
	}
	l.pos -= l.width
	l.curRune = eof
}

// curInput returns the current pending input of the lexer.
func (l *lexer) curInput() string {
	return l.input[l.start:l.pos]
}

// curPos returns the current position of the pending input
// of the lexer.
func (l *lexer) curPos() Pos {
	return Pos{Line: l.startLine, Column: l.startCol}
}

// pendingInput returns true, if pending input is available.
func (l *lexer) pendingInput() bool {
	return l.start < l.pos
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- Token{
		Type:  ILLEGAL,
		Value: fmt.Sprintf(format, args...),
		Pos:   l.curPos(),
	}
	return nil
}
