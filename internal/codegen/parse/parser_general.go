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

package parse

import (
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cloudfoundry/bytefmt"
	"github.com/desertbit/orbit/internal/codegen/ast"
)

// The returned name adheres to CamelCase.
// Returns ast.Err.
func (p *parser) expectName() (name string, line int, err error) {
	if p.empty() {
		err = ast.NewErr(p.prevLine, "expected name, but is missing")
		return
	}

	for _, r := range p.ct.Value {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
			err = ast.NewErr(p.ct.Line, "invalid char '%s' in name '%s'", string(r), p.ct.Value)
			return
		}
	}

	name = strings.Title(p.ct.Value)
	line = p.ct.Line

	// Advance to the next token.
	err = p.next()
	if err != nil {
		return
	}
	return
}

// Returns ast.Err.
func (p *parser) expectInt() (i int, err error) {
	if p.empty() {
		err = ast.NewErr(p.prevLine, "expected int, but is missing")
		return
	}

	i, err = strconv.Atoi(p.ct.Value)
	if err != nil {
		err = ast.NewErr(p.ct.Line, "expected int, %v", err)
		return
	}

	// Advance to the next token.
	err = p.next()
	if err != nil {
		return
	}
	return
}

// Consumes the current token, if it matches the symbol.
func (p *parser) checkSymbol(sym string) bool {
	if p.peekSymbol(sym) {
		// Advance to the next token.
		_ = p.next()
		return true
	}
	return false
}

// Consumes the current token.
// Returns ast.Err.
func (p *parser) expectSymbol(sym string) (err error) {
	if p.empty() {
		return ast.NewErr(p.prevLine, "expected '%s', but is missing", sym)
	}
	if p.ct.Value != sym {
		return ast.NewErr(p.ct.Line, "expected '%s', but got '%s'", sym, p.ct.Value)
	}

	// Advance to the next token.
	return p.next()
}

// Does not consume the current token.
func (p *parser) peekSymbol(sym string) bool {
	return p.value() == sym
}

// Consumes the current token.
// Returns ast.Err.
func (p *parser) expectDuration() (dur time.Duration, err error) {
	if p.empty() {
		err = ast.NewErr(p.prevLine, "expected time duration, but is missing")
		return
	}

	dur, err = time.ParseDuration(p.ct.Value)
	if err != nil {
		err = ast.NewErr(p.ct.Line, "expected time duration, %v", err)
		return
	}

	// Advance to the next token.
	err = p.next()
	if err != nil {
		return
	}

	return
}

// Consumes the current token.
// Returns ast.Err.
func (p *parser) expectByteSize() (size uint64, err error) {
	if p.empty() {
		err = ast.NewErr(p.prevLine, "expected byte size, but is missing")
		return
	}

	size, err = bytefmt.ToBytes(p.ct.Value)
	if err != nil {
		err = ast.NewErr(p.ct.Line, "invalid byte size, %v", err)
		return
	}

	// Advance to the next token.
	err = p.next()
	if err != nil {
		return
	}

	return
}
