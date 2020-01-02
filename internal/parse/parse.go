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
	"unicode"
)

const (
	tkService          = "service"
	tkType             = "type"
	tkEntryCall        = "call"
	tkEntryRevCall     = "revcall"
	tkEntryStream      = "stream"
	tkEntryParamStream = "stream"

	tkBraceL   = "{"
	tkBraceR   = "}"
	tkBracketL = "["
	tkBracketR = "]"
	tkBracket  = tkBracketL + tkBracketR
	tkParenL   = "("
	tkParenR   = ")"

	tkMap = "map"
)

func Parse(data string) (services []*Service, types []*StructType, err error) {
	// Tokenize the file.
	tks, err := tokenize(data)
	if err != nil {
		return
	}

	p, err := newParser(tks)
	if err != nil {
		return
	}

	services, err = p.parse()
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
		ck:    tks[0],
		types: make(map[string]*StructType),
	}
	return
}

func (p *parser) parse() (srvcs []*Service, err error) {
	for {
		// Either a service or a type can be declared top-level.
		if p.ck.value == tkService {
			// Expect service.
			var srvc *Service
			srvc, err = p.expectService()
			if err != nil {
				return
			}

			// Check, if the service has already been defined.
			for _, sr := range srvcs {
				if sr.Name == srvc.Name {
					err = &Error{msg: fmt.Sprintf("service '%s' declared twice"), line: p.ck.line}
					return
				}
			}

			srvcs = append(srvcs, srvc)
		} else if p.ck.value == tkType {
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
			err = &Error{msg: fmt.Sprintf("unknown top-level keyword '%s'", p.ck.value), line: p.ck.line}
			return
		}

		if !p.next() {
			// End.
			break
		}
	}
	return
}

func (p *parser) expectService() (srvc *Service, err error) {
	// Expecting the name.
	name, err := p.expectName()
	if err != nil {
		return
	}
	srvc = &Service{Name: name}

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

		srvc.Entries = append(srvc.Entries, e)
	}
}

func (p *parser) expectEntry() (e Entry, err error) {
	defer func() {
		var Err *Error
		if err != nil && !errors.As(err, &Err) {
			err = &Error{msg: err.Error(), line: p.ck.line}
		}
	}()

	// Expect the type.
	if !p.next() {
		err = &Error{msg: "expected entry type, but is missing", line: p.ck.line}
		return
	}
	if p.ck.value != tkEntryCall && p.ck.value != tkEntryRevCall && p.ck.value != tkEntryStream {
		err = fmt.Errorf("expected entry type, but got '%s'", p.ck.value)
		return
	}
	t := p.ck.value

	// Expect name.
	name, err := p.expectName()
	if err != nil {
		return
	}

	// If a stream, then no arguments.
	if t == tkEntryStream {
		e = &Stream{name: name}
		return
	}

	// Check for arguments.
	// Check '('.
	var args, ret *EntryParam
	if p.checkSymbol(tkParenL) {
		args, err = p.expectEntryParam(name, "Args")
		if err != nil {
			return
		} else if args != nil && p.checkSymbol(tkParenL) {
			// arguments provided, check for returns.
			ret, err = p.expectEntryParam(name, "Ret")
			if err != nil {
				return
			}
		}
	}

	// Create entry based on type.
	if t == tkEntryCall {
		e = &Call{name: name, Args: args, Ret: ret}
	} else {
		e = &RevCall{name: name, Args: args, Ret: ret}
	}
	return
}

// Opening parenthesis must already be consumed!
// Closing parenthesis is consumed in this method.
// Returns Error.
func (p *parser) expectEntryParam(entryName, inPlaceSuffix string) (ep *EntryParam, err error) {
	// Check ')' (empty params).
	if p.checkSymbol(tkParenR) {
		return
	}

	// Check 'stream'.
	isStream := p.checkSymbol(tkEntryParamStream)

	// Check '{' (in-place type).
	// Otherwise, expect struct type reference.
	var t *StructType
	if p.checkSymbol(tkBraceL) {
		t, err = p.expectStructType(entryName + inPlaceSuffix)
	} else {
		t, err = p.expectStructTypeRef()
	}
	if err != nil {
		return
	}

	// Expect ')'.
	err = p.expectSymbol(tkParenR)
	if err != nil {
		return
	}

	ep = &EntryParam{Type: t, IsStream: isStream}
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

// Returns Error.
func (p *parser) expectName() (name string, err error) {
	defer func() {
		if err != nil {
			err = &Error{msg: err.Error(), line: p.ck.line}
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
		err = &Error{msg: fmt.Sprintf("expected base type, but got '%s'", p.ck.value), line: p.ck.line}
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
		err = &Error{msg: fmt.Sprintf("type '%s' declared twice", name), line: p.ck.line}
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

func (p *parser) checkSymbol(sym string) bool {
	if p.peek() == sym {
		// Consume token.
		_ = p.next()
		return true
	}
	return false
}

// Returns Error.
func (p *parser) expectSymbol(sym string) (err error) {
	defer func() {
		if err != nil {
			err = &Error{msg: err.Error(), line: p.ck.line}
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

func (p *parser) next() (ok bool) {
	if p.ti == len(p.tks)-1 {
		return false
	}
	p.ti++
	p.ck = p.tks[p.ti]
	return true
}

func (p *parser) peek() (v string) {
	if p.ti == len(p.tks)-1 {
		return
	}
	v = p.tks[p.ti+1].value
	return
}
