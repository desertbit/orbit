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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type parser struct {
	tks []*token
	ti  int
	ct  *token

	// Stores all services by their name.
	srvcs        map[string]*Service
	srvcsStructs map[string][]*StructType

	// Stores all types by their name.
	types map[string]*Type

	// Stores all errors by their name.
	errors map[string]*Error
}

func newParser(tks []*token) (p *parser, err error) {
	if len(tks) == 0 {
		err = errors.New("empty tokens")
		return
	}
	p = &parser{
		tks:          tks,
		ti:           -1,
		ct:           tks[0],
		srvcs:        make(map[string]*Service),
		srvcsStructs: make(map[string][]*StructType),
		types:        make(map[string]*Type),
		errors:       make(map[string]*Error),
	}
	return
}

// Can only be called once per parser!
func (p *parser) parse() (srvcs []*Service, types []*Type, errors []*Error, err error) {
	for {
		// Either an error, service or a type can be declared top-level.
		if p.empty() {
			break
		} else if p.checkSymbol(tkErrors) {
			// Expect error definitions.
			err = p.expectErrors("")
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkService) {
			// Expect service.
			err = p.expectService()
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkType) {
			// Expect type.
			_, err = p.expectType("", "")
			if err != nil {
				return
			}
		} else {
			err = &Err{msg: fmt.Sprintf("unknown top-level keyword '%s'", p.ct.value), line: p.ct.line}
			return
		}
	}

	// Extract the return values from the parser.
	srvcs = make([]*Service, 0, len(p.srvcs))
	for _, srvc := range p.srvcs {
		srvcs = append(srvcs, srvc)
	}
	types = make([]*Type, 0, len(p.types))
	for _, t := range p.types {
		types = append(types, t)
	}
	errors = make([]*Error, 0, len(p.errors))
	for _, e := range p.errors {
		errors = append(errors, e)
	}

	return
}

// The returned name adheres to CamelCase.
// Returns Err.
func (p *parser) expectName() (name string, err error) {
	defer func() {
		if err != nil {
			err = &Err{msg: err.Error(), line: p.ct.line}
		}
	}()

	if !p.next() || p.ct.value == "" {
		err = errors.New("expected name, but is missing")
		return
	}

	for _, r := range p.ct.value {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
			err = fmt.Errorf("invalid char '%s' in name '%s'", string(r), p.ct.value)
			return
		}
	}

	name = strings.Title(p.ct.value)
	return
}

func (p *parser) expectInt() (i int, err error) {
	defer func() {
		if err != nil {
			err = &Err{msg: err.Error(), line: p.ct.line}
		}
	}()

	if !p.next() || p.ct.value == "" {
		err = errors.New("expected int, but is missing")
		return
	}

	return strconv.Atoi(p.ct.value)
}

// Consumes the current token, if it matches the symbol.
func (p *parser) checkSymbol(sym string) bool {
	if p.peek() == sym {
		// Consume token.
		_ = p.next()
		return true
	}
	return false
}

// Consumes the current token.
// Returns Err.
func (p *parser) expectSymbol(sym string) (err error) {
	defer func() {
		if err != nil {
			err = &Err{msg: err.Error(), line: p.ct.line}
		}
	}()

	if !p.next() {
		return fmt.Errorf("expected '%s', but is missing", sym)
	}
	if p.ct.value != sym {
		return fmt.Errorf("expected '%s', but got '%s'", sym, p.ct.value)
	}
	return
}

// Does not consume the current token.
func (p *parser) peekSymbol(sym string) bool {
	return p.peek() == sym
}

func (p *parser) empty() bool {
	return p.ti >= len(p.tks)-1
}

func (p *parser) next() (ok bool) {
	if p.empty() {
		return false
	}
	p.ti++
	p.ct = p.tks[p.ti]
	return true
}

func (p *parser) peek() (v string) {
	if p.empty() {
		return
	}
	v = p.tks[p.ti+1].value
	return
}
