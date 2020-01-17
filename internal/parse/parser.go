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
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type parser struct {
	tks []*token
	ti  int
	ct  *token
	lt  *token

	srvcs map[string]*Service
	types map[string]*Type
	errs  map[string]*Error
}

func newParser(tks []*token) (p *parser, err error) {
	p = &parser{
		tks: tks,

		srvcs: make(map[string]*Service),
		types: make(map[string]*Type),
		errs:  make(map[string]*Error),
	}
	if len(tks) > 0 {
		p.ct = tks[0]
		p.lt = p.ct
	}
	return
}

// Can only be called once per parser!
func (p *parser) parse() (srvcs []*Service, types []*Type, errors []*Error, err error) {
	var (
		errs []*Error
		srvc *Service
		t    *Type
	)

	for {
		if p.empty() {
			// No more tokens left.
			break
		} else if p.checkSymbol(tkErrors) {
			// Expect global error definitions.
			errs, err = p.expectErrors("")
			if err != nil {
				return
			}

			for _, e := range errs {
				p.errs[e.Name] = e
			}
		} else if p.checkSymbol(tkService) {
			// Expect service.
			srvc, err = p.expectService()
			if err != nil {
				return
			}

			p.srvcs[srvc.Name] = srvc
		} else if p.checkSymbol(tkType) {
			// Expect global type.
			t, _, err = p.expectType("", "")
			if err != nil {
				return
			}

			p.types[t.Name] = t
		} else {
			err = &Err{
				msg:  fmt.Sprintf("unknown top-level keyword '%s'", p.ct.value),
				line: p.ct.line,
			}
			return
		}
	}

	srvcs = make([]*Service, 0, len(p.srvcs))
	for _, srvc := range p.srvcs {
		srvcs = append(srvcs, srvc)
	}
	types = make([]*Type, 0, len(p.types))
	for _, t := range p.types {
		types = append(types, t)
	}
	errors = make([]*Error, 0, len(p.errs))
	for _, e := range p.errs {
		errors = append(errors, e)
	}

	return
}

// The returned name adheres to CamelCase.
// Returns Err.
func (p *parser) expectName() (name string, err error) {
	if p.empty() {
		err = &Err{msg: "expected name, but is missing", line: p.ct.line}
		return
	}

	for _, r := range p.ct.value {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
			err = &Err{
				msg:  fmt.Sprintf("invalid char '%s' in name '%s'", string(r), p.ct.value),
				line: p.lt.line,
			}
			return
		}
	}

	name = strings.Title(p.ct.value)

	// Consume token.
	p.consume()
	return
}

// Returns Err.
func (p *parser) expectInt() (i int, err error) {
	if p.empty() {
		err = &Err{msg: "expected int, but is missing", line: p.ct.line}
		return
	}

	i, err = strconv.Atoi(p.ct.value)
	if err != nil {
		err = &Err{msg: err.Error(), line: p.ct.line}
		return
	}

	// Consume token.
	p.consume()
	return
}

// Consumes the current token, if it matches the symbol.
func (p *parser) checkSymbol(sym string) bool {
	if p.peek() == sym {
		// Consume token.
		p.consume()
		return true
	}
	return false
}

// Consumes the current token.
// Returns Err.
func (p *parser) expectSymbol(sym string) (err error) {
	if p.empty() {
		return &Err{
			msg:  fmt.Sprintf("expected '%s', but is missing", sym),
			line: p.lt.line,
		}
	}
	if p.ct.value != sym {
		return &Err{
			msg:  fmt.Sprintf("expected '%s', but got '%s'", sym, p.ct.value),
			line: p.ct.line,
		}
	}
	// Consume the token.
	p.consume()
	return
}

// Does not consume the current token.
func (p *parser) peekSymbol(sym string) bool {
	return p.peek() == sym
}

func (p *parser) empty() bool {
	return p.ti >= len(p.tks)
}

func (p *parser) consume() {
	p.lt = p.ct
	p.ti++
	if !p.empty() {
		p.ct = p.tks[p.ti]
	}
}

func (p *parser) peek() string {
	if p.empty() {
		return ""
	}
	return p.ct.value
}
