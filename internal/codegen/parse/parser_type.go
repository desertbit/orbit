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

	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/token"
)

// Returns ast.Err.
func (p *parser) expectType() (err error) {
	// Expect name.
	name, line, err := p.expectName()
	if err != nil {
		return
	}

	// Expect type definition.
	return p.expectInlineType(name, line)
}

func (p *parser) expectInlineType(name string, line int) (err error) {
	t := &ast.Type{Name: name, Line: line}

	// Expect '{'.
	err = p.expectSymbol(token.BraceL)
	if err != nil {
		return
	}

	// Expect type fields.
	for {
		// Check end.
		if p.checkSymbol(token.BraceR) {
			break
		}

		tf := &ast.TypeField{}

		// Expect name.
		tf.Name, tf.Line, err = p.expectName()
		if err != nil {
			return
		}

		// Expect type.
		tf.DataType, err = p.expectDataType()
		if err != nil {
			return
		}

		// Check for a validation tag definition.
		if p.checkSymbol(token.SingQuote) {
			tf.ValTag, err = p.expectValTag()
			if err != nil {
				return
			}
		}

		t.Fields = append(t.Fields, tf)
	}

	// Add to types.
	p.types = append(p.types, t)

	return
}

// Returns Err.
func (p *parser) expectValTag() (valTag string, err error) {
	// Expect the tag.
	if p.empty() {
		err = ast.NewErr(p.prevLine, "expected validation tag, but is missing")
		return
	}
	valTag = p.ct.Value

	// Advance to the next token.
	err = p.next()
	if err != nil {
		return
	}

	// Check for ending single quote.
	err = p.expectSymbol(token.SingQuote)
	if err != nil {
		return
	}
	return
}

// Returns ast.Err.
func (p *parser) expectDataType() (t ast.DataType, err error) {
	if p.checkSymbol(token.BracketL) && p.checkSymbol(token.BracketR) {
		// Array type.
		t, err = p.expectArrType()
		if err != nil {
			return
		}
	} else if p.checkSymbol(tkMap) {
		// Map type.
		t, err = p.expectMapType()
		if err != nil {
			return
		}
	} else {
		// Struct or Base type.
		var ok bool
		t, ok = p.checkBaseType()
		if !ok {
			// Expect any type.
			t, err = p.expectAnyType()
			if err != nil {
				return
			}
		}
	}
	return
}

func (p *parser) checkBaseType() (b *ast.BaseType, ok bool) {
	val := p.value()
	switch val {
	case ast.TypeByte, ast.TypeString, ast.TypeTime,
		ast.TypeBool,
		ast.TypeInt, ast.TypeInt8, ast.TypeInt16, ast.TypeInt32, ast.TypeInt64,
		ast.TypeUInt, ast.TypeUInt8, ast.TypeUInt16, ast.TypeUInt32, ast.TypeUInt64,
		ast.TypeFloat32, ast.TypeFloat64:
		ok = true
		b = &ast.BaseType{DataType: val, Line: p.line()}

		// Consume token.
		_ = p.next()
	}
	return
}

// Returns ast.Err.
func (p *parser) expectMapType() (m *ast.MapType, err error) {
	m = &ast.MapType{Line: p.prevLine}

	// Expect '['.
	err = p.expectSymbol(token.BracketL)
	if err != nil {
		return
	}

	// Expect key type, must be base type.
	m.Key, err = p.expectBaseType()
	if err != nil {
		return
	}

	// Expect ']'.
	err = p.expectSymbol(token.BracketR)
	if err != nil {
		return
	}

	// Expect value type, can be anything.
	m.Value, err = p.expectDataType()
	if err != nil {
		return
	}

	return
}

// Returns ast.Err.
func (p *parser) expectArrType() (a *ast.ArrType, err error) {
	a = &ast.ArrType{Line: p.prevLine}

	// Expect any type.
	a.Elem, err = p.expectDataType()
	if err != nil {
		return
	}

	return
}

// Returns ast.Err.
func (p *parser) expectBaseType() (b *ast.BaseType, err error) {
	b, ok := p.checkBaseType()
	if !ok {
		err = ast.NewErr(p.line(), "expected base type, but got '%s'", p.value())
		return
	}
	return
}

// Returns ast.Err.
func (p *parser) expectAnyType() (a *ast.AnyType, err error) {
	a = &ast.AnyType{}

	// Expect name.
	a.NamePrv, a.Line, err = p.expectName()
	if err != nil {
		err = fmt.Errorf("invalid type: %w", err)
		return
	}

	return
}
