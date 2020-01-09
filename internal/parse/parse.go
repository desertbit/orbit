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

const (
	tkErrors  = "errors"
	tkService = "service"
	tkType    = "type"

	tkEntryAsync     = "async"
	tkEntryCall      = "call"
	tkEntryRevCall   = "revcall"
	tkEntryStream    = "stream"
	tkEntryRevStream = "revstream"

	tkEqual    = "="
	tkBraceL   = "{"
	tkBraceR   = "}"
	tkBracketL = "["
	tkBracketR = "]"
	tkBracket  = tkBracketL + tkBracketR
	tkParenL   = "("
	tkParenR   = ")"

	tkMap = "map"
)

func Parse(data string) (errors []*Error, services []*Service, types []*StructType, err error) {
	// Tokenize the file.
	tks, err := tokenize(data)
	if err != nil {
		return
	}

	p, err := newParser(tks)
	if err != nil {
		return
	}

	errors, services, err = p.parse()
	if err != nil {
		return
	}

	types = make([]*StructType, 0, len(p.types))
	for _, t := range p.types {
		types = append(types, t)
	}

	return
}

type parser struct {
	tks []*token
	ti  int
	ck  *token

	// Stores all types by their name.
	types map[string]*StructType
}

func newParser(tks []*token) (p *parser, err error) {
	if len(tks) == 0 {
		err = errors.New("empty tokens")
		return
	}
	p = &parser{
		tks:   tks,
		ti:    -1,
		ck:    tks[0],
		types: make(map[string]*StructType),
	}
	return
}

func (p *parser) parse() (errors []*Error, srvcs []*Service, err error) {
	for {
		// Either an error, service or a type can be declared top-level.
		if p.checkSymbol(tkErrors) {
			// Expect error definitions.
			var newErrors []*Error
			newErrors, err = p.expectErrors()
			if err != nil {
				return
			}

			// Check for duplicates.
			for _, ne := range newErrors {
				for _, e := range errors {
					if ne.Name == e.Name {
						err = &Err{msg: fmt.Sprintf("errors '%s' declared twice", ne.Name), line: p.ck.line}
						return
					} else if ne.ID == e.ID {
						err = &Err{
							msg:  fmt.Sprintf("errors '%s' and '%s' share same identifier", ne.Name, e.Name),
							line: p.ck.line,
						}
						return
					}
				}
			}

			errors = append(errors, newErrors...)
		} else if p.checkSymbol(tkService) {
			// Expect service.
			var srvc *Service
			srvc, err = p.expectService()
			if err != nil {
				return
			}

			// Check, if the service has already been defined.
			for _, sr := range srvcs {
				if sr.Name == srvc.Name {
					err = &Err{msg: fmt.Sprintf("service '%s' declared twice", sr.Name), line: p.ck.line}
					return
				}
			}

			srvcs = append(srvcs, srvc)
		} else if p.checkSymbol(tkType) {
			// Expect name.
			var name string
			name, err = p.expectName()
			if err != nil {
				return
			}

			// Expect '{'.
			err = p.expectSymbol(tkBraceL)
			if err != nil {
				return
			}

			// Expect struct type.
			_, err = p.expectStructType(name)
			if err != nil {
				return
			}
		} else {
			if !p.empty() {
				err = &Err{msg: fmt.Sprintf("unknown top-level keyword '%s'", p.ck.value), line: p.ck.line}
			}
			return
		}
	}
}

func (p *parser) expectErrors() (errors []*Error, err error) {
	// Expect "{".
	err = p.expectSymbol(tkBraceL)
	if err != nil {
		return
	}

	for {
		// Check for end of service.
		if p.checkSymbol(tkBraceR) {
			return
		}

		// Expect name.
		var name string
		name, err = p.expectName()
		if err != nil {
			return
		}
		// Ensure CamelCase.
		name = strings.Title(name)

		// Expect "=".
		err = p.expectSymbol(tkEqual)
		if err != nil {
			return
		}

		// Expect identifier.
		var id int
		id, err = p.expectInt()
		if err != nil {
			return
		}

		errors = append(errors, &Error{Name: name, ID: id})
	}
}

func (p *parser) expectService() (srvc *Service, err error) {
	// Expecting the name.
	name, err := p.expectName()
	if err != nil {
		return
	}
	// Ensure CamelCase.
	srvc = &Service{Name: strings.Title(name)}

	// Expecting "{"
	err = p.expectSymbol(tkBraceL)
	if err != nil {
		return
	}

	for {
		// Check for end of service.
		if p.checkSymbol(tkBraceR) {
			return
		}

		// Expect entry.
		var e Entry
		e, err = p.expectEntry()
		if err != nil {
			return
		}

		// Check for duplicates.
		for _, en := range srvc.Entries {
			if en.NamePrv() == e.NamePrv() {
				err = &Err{
					msg:  fmt.Sprintf("entry '%s' declared twice in service '%s'", e.NamePrv(), srvc.Name),
					line: p.ck.line,
				}
			}
		}

		srvc.Entries = append(srvc.Entries, e)
	}
}

func (p *parser) expectEntry() (e Entry, err error) {
	defer func() {
		var pErr *Err
		if err != nil && !errors.As(err, &pErr) {
			err = &Err{msg: pErr.Error(), line: p.ck.line}
		}
	}()

	// Check, if async.
	async := p.checkSymbol(tkEntryAsync)

	// Expect the type.
	if !p.next() {
		err = errors.New("expected entry type, but is missing")
		return
	}
	t := p.ck.value

	// Validate type.
	if t != tkEntryCall && t != tkEntryRevCall && t != tkEntryStream && t != tkEntryRevStream {
		err = fmt.Errorf("expected entry type, but got '%s'", p.ck.value)
		return
	}

	// Async is only allowed for calls and revcalls.
	if async && (t == tkEntryStream || t == tkEntryRevStream) {
		err = errors.New("async not allowed here")
		return
	}

	// Expect name.
	name, err := p.expectName()
	if err != nil {
		return
	}
	// Ensure CamelCase.
	name = strings.Title(name)

	// Check for arguments.
	var args, ret *EntryParam
	args, empty, err := p.checkEntryParam(name, "Args")
	if err != nil {
		return
	} else if empty || args != nil {
		ret, _, err = p.checkEntryParam(name, "Ret")
		if err != nil {
			return
		}
	}

	// Create entry based on type.
	if t == tkEntryCall || t == tkEntryRevCall {
		e = &Call{name: name, rev: t == tkEntryRevCall, Async: async, Args: args, Ret: ret}
	} else {
		e = &Stream{name: name, rev: t == tkEntryRevStream, Args: args, Ret: ret}
	}
	return
}

// Opening parenthesis must already be consumed!
// Closing parenthesis is consumed in this method.
// Returns Err.
func (p *parser) checkEntryParam(entryName, inPlaceSuffix string) (ep *EntryParam, empty bool, err error) {
	// Check '('.
	if !p.checkSymbol(tkParenL) {
		return
	}

	// Check ')' (empty params).
	if p.checkSymbol(tkParenR) {
		empty = true
		return
	}
	ep = &EntryParam{}

	// Check '{' (in-place type).
	// Otherwise, expect struct type reference.
	if p.checkSymbol(tkBraceL) {
		ep.Type, err = p.expectStructType(entryName + inPlaceSuffix)
	} else {
		ep.Type, err = p.expectStructTypeRef()
	}
	if err != nil {
		return
	}

	// Expect ')'.
	err = p.expectSymbol(tkParenR)
	if err != nil {
		return
	}

	return
}

func (p *parser) expectType() (t Type, err error) {
	var ts string
	defer func() {
		if err != nil {
			err = fmt.Errorf("parsing %s type: %w", ts, err)
		}
	}()

	if p.checkSymbol(tkBracket) {
		// Array type.
		ts = "array"

		t, err = p.expectArrType()
		if err != nil {
			return
		}
	} else if p.checkSymbol(tkMap) {
		// Map type.
		ts = tkMap

		t, err = p.expectMapType()
		if err != nil {
			return
		}
	} else {
		ts = "Struct or Base"

		// Struct or Base type.
		var ok bool
		t, ok = p.checkBaseType()
		if !ok {
			// Expect struct type.
			t, err = p.expectStructTypeRef()
			if err != nil {
				return
			}
		}
		return
	}
	return
}

func (p *parser) checkBaseType() (b *BaseType, ok bool) {
	switch p.peek() {
	case TypeByte, TypeString, TypeTime,
		TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64,
		TypeUInt, TypeUInt8, TypeUInt16, TypeUInt32, TypeUInt64,
		TypeFloat32, TypeFloat64:
		// Consume token.
		_ = p.next()
		ok = true
		b = &BaseType{dataType: p.ck.value}
	}
	return
}

func (p *parser) expectMapType() (m *MapType, err error) {
	// Expect '['.
	err = p.expectSymbol(tkBracketL)
	if err != nil {
		return
	}

	// Expect key type, must be base type.
	var key *BaseType
	key, err = p.expectBaseType()
	if err != nil {
		return
	}

	// Expect ']'.
	err = p.expectSymbol(tkBracketR)
	if err != nil {
		return
	}

	// Expect value type, can be anything.
	var value Type
	value, err = p.expectType()
	if err != nil {
		return
	}

	m = &MapType{Key: key, Value: value}
	return
}

func (p *parser) expectArrType() (a *ArrType, err error) {
	// Expect any type.
	t, err := p.expectType()
	if err != nil {
		return
	}

	a = &ArrType{ElemType: t}
	return
}

func (p *parser) expectBaseType() (b *BaseType, err error) {
	b, ok := p.checkBaseType()
	if !ok {
		err = &Err{msg: fmt.Sprintf("expected base type, but got '%s'", p.ck.value), line: p.ck.line}
		return
	}
	return
}

func (p *parser) expectStructType(name string) (s *StructType, err error) {
	// Check, if the struct has already been declared.
	// References have the fields of the struct not filled yet.
	s, ok := p.types[name]
	if !ok {
		// New type.
		s = &StructType{Name: name}
		p.types[name] = s
	} else if s.Fields != nil {
		err = &Err{msg: fmt.Sprintf("type '%s' declared twice", name), line: p.ck.line}
		return
	}

	// Parse fields of struct.
	for {
		// Check end.
		if p.checkSymbol(tkBraceR) {
			return
		}

		f := &StructField{}

		// Expect name.
		f.Name, err = p.expectName()
		if err != nil {
			return
		}
		f.Name = strings.Title(f.Name)

		// Expect type.
		f.Type, err = p.expectType()
		if err != nil {
			return
		}

		s.Fields = append(s.Fields, f)
	}
}

func (p *parser) expectStructTypeRef() (s *StructType, err error) {
	// Expect name.
	name, err := p.expectName()
	if err != nil {
		err = fmt.Errorf("invalid struct type: %w", err)
		return
	}

	// Check, if type already exists.
	var ok bool
	if s, ok = p.types[name]; ok {
		return
	}

	// Add as new type.
	s = &StructType{Name: name}
	p.types[name] = s
	return
}

// Returns Err.
func (p *parser) expectName() (name string, err error) {
	defer func() {
		if err != nil {
			err = &Err{msg: err.Error(), line: p.ck.line}
		}
	}()

	if !p.next() || p.ck.value == "" {
		err = errors.New("expected name, but is missing")
		return
	}

	for _, r := range p.ck.value {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
			err = fmt.Errorf("invalid char '%s' in name '%s'", string(r), p.ck.value)
			return
		}
	}

	name = p.ck.value
	return
}

func (p *parser) expectInt() (i int, err error) {
	defer func() {
		if err != nil {
			err = &Err{msg: err.Error(), line: p.ck.line}
		}
	}()

	if !p.next() || p.ck.value == "" {
		err = errors.New("expected int, but is missing")
		return
	}

	return strconv.Atoi(p.ck.value)
}

func (p *parser) checkSymbol(sym string) bool {
	if p.peek() == sym {
		// Consume token.
		_ = p.next()
		return true
	}
	return false
}

// Returns Err.
func (p *parser) expectSymbol(sym string) (err error) {
	defer func() {
		if err != nil {
			err = &Err{msg: err.Error(), line: p.ck.line}
		}
	}()

	if !p.next() {
		return fmt.Errorf("expected '%s', but is missing", sym)
	}
	if p.ck.value != sym {
		return fmt.Errorf("expected '%s', but got '%s'", sym, p.ck.value)
	}
	return
}

func (p *parser) empty() bool {
	return p.ti >= len(p.tks)-1
}

func (p *parser) next() (ok bool) {
	if p.empty() {
		return false
	}
	p.ti++
	p.ck = p.tks[p.ti]
	return true
}

func (p *parser) peek() (v string) {
	if p.empty() {
		return
	}
	v = p.tks[p.ti+1].value
	return
}
