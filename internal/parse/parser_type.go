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
)

// Returns Err.
func (p *parser) expectType(srvcName, name string) (t *Type, sts []*StructType, err error) {
	// Expect name if empty and prepend the prefix.
	if name == "" {
		name, err = p.expectName()
		if err != nil {
			return
		}
	}
	name = srvcName + name

	t = &Type{Name: name, line: p.ct.line}

	// Expect '{'.
	err = p.expectSymbol(tkBraceL)
	if err != nil {
		return
	}

	// Expect type fields.
	t.Fields, sts, err = p.expectTypeFields()
	if err != nil {
		return
	}

	// Validate.
	err = p.validateType(t)
	if err != nil {
		return
	}

	return
}

// Returns Err.
func (p *parser) validateType(t *Type) (err error) {
	defer func() {
		if err != nil {
			err = &Err{msg: err.Error(), line: t.line}
		}
	}()

	// Check for duplicate names.
	if _, ok := p.types[t.Name]; ok {
		err = fmt.Errorf("type '%s' declared twice", t.Name)
		return
	}

	return
}

// Returns Err.
func (p *parser) expectTypeFields() (tfs []*TypeField, sts []*StructType, err error) {
	// Parse fields of type.
	for {
		// Check end.
		if p.checkSymbol(tkBraceR) {
			return
		}

		tf := &TypeField{}

		// Expect name.
		tf.Name, err = p.expectName()
		if err != nil {
			return
		}

		// Expect type.
		tf.DataType, err = p.expectDataType()
		if err != nil {
			return
		}

		// In case the data type contains a struct type, add it to the structs.
		s := containsStruct(tf.DataType)
		if s != nil {
			sts = append(sts, s)
		}

		tfs = append(tfs, tf)
	}
}

// Returns Err.
func (p *parser) expectDataType() (t DataType, err error) {
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
			t, err = p.expectStructType()
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
		TypeBool,
		TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64,
		TypeUInt, TypeUInt8, TypeUInt16, TypeUInt32, TypeUInt64,
		TypeFloat32, TypeFloat64:
		ok = true
		b = &BaseType{dataType: p.ct.value}

		// Consume token.
		p.consume()
	}
	return
}

// Returns Err.
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
	var value DataType
	value, err = p.expectDataType()
	if err != nil {
		return
	}

	m = &MapType{Key: key, Value: value}
	return
}

// Returns Err.
func (p *parser) expectArrType() (a *ArrType, err error) {
	// Expect any type.
	t, err := p.expectDataType()
	if err != nil {
		return
	}

	a = &ArrType{Elem: t}
	return
}

// Returns Err.
func (p *parser) expectBaseType() (b *BaseType, err error) {
	b, ok := p.checkBaseType()
	if !ok {
		err = &Err{msg: fmt.Sprintf("expected base type, but got '%s'", p.ct.value), line: p.ct.line}
		return
	}
	return
}

// Returns Err.
func (p *parser) expectStructType() (s *StructType, err error) {
	// Expect name.
	name, err := p.expectName()
	if err != nil {
		err = fmt.Errorf("invalid struct type: %w", err)
		return
	}

	s = &StructType{Name: name}
	return
}
