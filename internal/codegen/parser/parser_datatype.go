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

package parser

import (
	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/lexer"
)

func (p *parser) expectTypeDefinition() ([]*ast.TypeField, error) {
	// Orbit file example:
	/*
		<type-declaration or inline-type>
			name    string `validate:"required,min=1"`
			age     int    `validate:"required,min=1,max=155"`
			locale  string `validate:"required,len=5"`
			address string `validate:"omitempty"`
		}
	*/

	// Type Fields.
	var tfs []*ast.TypeField
	for !p.checkToken(lexer.RBRACE) {
		tf := &ast.TypeField{Pos: p.tk.Pos}

		// Identifier.
		var err error
		tf.Name, err = p.expectIdent()
		if err != nil {
			return nil, err
		}

		// Data type.
		tf.DataType, err = p.expectDataType()
		if err != nil {
			return nil, err
		}

		// Optional struct tag.
		if p.checkToken(lexer.RAWSTRING) {
			tf.StructTag = p.tk.Value
		}

		// Add type field to type.
		tfs = append(tfs, tf)
	}

	return tfs, nil
}

func (p *parser) expectDataType() (ast.DataType, error) {
	if p.checkToken(lexer.LBRACK) && p.checkToken(lexer.RBRACK) {
		return p.expectArrType()
	} else if p.checkToken(lexer.MAP) {
		return p.expectMapType()
	} else {
		return p.expectAnyType()
	}
}

func (p *parser) expectMapType() (*ast.MapType, error) {
	mt := &ast.MapType{Pos: p.tk.Pos}

	// '['.
	err := p.expectToken(lexer.LBRACK)
	if err != nil {
		return nil, err
	}

	// Key type.
	mt.Key, err = p.expectAnyType()
	if err != nil {
		return nil, err
	}

	// ']'.
	err = p.expectToken(lexer.RBRACK)
	if err != nil {
		return nil, err
	}

	// Value type.
	mt.Value, err = p.expectDataType()
	if err != nil {
		return nil, err
	}

	return mt, nil
}

func (p *parser) expectArrType() (*ast.ArrType, error) {
	at := &ast.ArrType{Pos: p.tk.Pos}

	// Expect any type.
	var err error
	at.Elem, err = p.expectDataType()
	if err != nil {
		return nil, err
	}

	return at, nil
}

func (p *parser) expectAnyType() (*ast.AnyType, error) {
	at := &ast.AnyType{Pos: p.tk.Pos}

	// Identifier.
	var err error
	at.Name, err = p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Ensure private name is lowercase.
	return at, nil
}
