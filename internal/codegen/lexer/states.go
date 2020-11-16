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

package lexer

import (
	"strconv"
	"time"
	"unicode"

	"code.cloudfoundry.org/bytefmt"
)

func lexTokenStart(l *lexer) stateFn {
	var (
		r rune
		t TokenType
	)

	// Discard leading whitespace.
	for r = l.next(); r != eof; r = l.next() {
		if !unicode.IsSpace(r) {
			l.backup()
			break
		}
	}
	l.ignore()

	// Retrieve the first rune and check,
	// if it is a special symbol we must
	// emit a token for separately.
	r = l.next()

	// Number.
	if r == hyphen || unicode.IsDigit(r) {
		l.backup()
		return lexNumber
	}

	// Delimiter.
	t = toDelimTokenType(r)
	if t != ILLEGAL {
		l.emit(t)
		return lexTokenStart
	}

	// Raw string.
	if r == backtick {
		l.ignore()
		return lexRawString
	}

	// Comments.
	if r == slash {
		// May be a comment.
		nr := l.next()
		if nr == slash {
			return lexLineComment
		} else if nr == asterisk {
			return lexBlockComment
		}
		// No comment.
		l.backup()
	}

	return lexTokenFindEnd
}

func lexTokenFindEnd(l *lexer) stateFn {
	// Find next boundary.
	for r := l.next(); r != eof; r = l.next() {
		if unicode.IsSpace(r) || isDelim(r) || r == backtick {
			l.backup()
			break
		}
	}

	// Valid EOF.
	if !l.pendingInput() {
		return nil
	}

	// Check for keywords.
	t := toKeywordTokenType(l.curInput())
	if t != ILLEGAL {
		l.emit(t)
	} else {
		// Must be identifier.
		l.emit(IDENT)
	}

	return lexTokenStart
}

func lexRawString(l *lexer) stateFn {
	// Collect input until next rquote.
	// Newlines are an error for raw strings.
	for r := l.next(); r != eof; r = l.next() {
		if r == newline {
			return l.errorf("unexpected newline in raw string literal")
		} else if r == backtick {
			// End of string.
			// Remove rquote from input.
			l.backup()
			// Publish input.
			l.emit(RAWSTRING)
			// Ignore rquote.
			l.next()
			l.ignore()
			return lexTokenStart
		}
	}

	// Invalid EOF.
	return l.errorf("unterminated raw string literal")
}

func lexNumber(l *lexer) stateFn {
	// Collect input until next space or eof.
	var (
		r          rune
		onlyDigits = true
	)
	for r = l.next(); r != eof; r = l.next() {
		if unicode.IsSpace(r) {
			// Do not include space in input.
			l.backup()
			break
		}

		if r != hyphen && !unicode.IsDigit(r) {
			onlyDigits = false
		} else if !onlyDigits {
			// After a non-digit, no digit can follow.
			return l.errorf("invalid number literal: '%s'", l.curInput())
		}
	}

	if !l.pendingInput() && r == eof {
		// Invalid EOF.
		return l.errorf("unterminated number literal")
	}
	ci := l.curInput()

	if onlyDigits {
		// Integer, use strconv package.
		_, err := strconv.Atoi(ci)
		if err != nil {
			return l.errorf("invalid number literal: %v", err)
		}

		// Publish int.
		l.emit(INT)
		return lexTokenStart
	}

	// If the number contains non-digits, try to convert it to a byte size / duration.
	// Try faster byte size first.
	_, err := bytefmt.ToBytes(ci)
	if err == nil {
		// Valid.
		l.emit(BYTESIZE)
		return lexTokenStart
	}

	// Try duration.
	_, err = time.ParseDuration(ci)
	if err == nil {
		// Valid.
		l.emit(DURATION)
		return lexTokenStart
	}

	// Invalid.
	return l.errorf("invalid number literal: '%s'", ci)
}

func lexLineComment(l *lexer) stateFn {
	// Consume tokens until a newline or eof is encountered.
	for r := l.next(); r != eof && r != newline; r = l.next() {
	}
	l.ignore()
	return lexTokenStart
}

func lexBlockComment(l *lexer) stateFn {
	// Consume tokens until block comment end or eof is encountered.
	for r := l.next(); r != eof; r = l.next() {
		if r == asterisk {
			r = l.next()
			if r == slash {
				l.ignore()
				return lexTokenStart
			}
		}
	}

	// Valid EOF.
	l.ignore()
	return nil
}
