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

func (p *parser) run() (*ast.File, error) {
	// Create the final ast.File we want to return.
	f := &ast.File{}

	// Retrieve the next token.
	for {
		err := p.next()
		if err != nil {
			return nil, err
		} else if p.tk.Type == lexer.EOF {
			// Valid EOF.
			return f, nil
		} else if !p.tk.IsKeyword() {
			// Expected a keyword.
			return nil, p.errorf("expected top-level keyword, found %s", p.tk.Value)
		}

		// Found a keyword.
		// Now check, if it is a top-level one.
		switch p.tk.Type {
		case lexer.VERSION:
			err = p.parseVersion(f)
		case lexer.ENUM:
			err = p.parseEnum(f)
		case lexer.ERRORS:
			err = p.parseErrors(f)
		case lexer.TYPE:
			err = p.parseTypeDeclaration(f)
		case lexer.SERVICE:
			err = p.parseService(f)
		default:
			err = p.errorf("invalid top-level keyword %s", p.tk.Value)
		}
		if err != nil {
			return nil, err
		}
	}
}

func (p *parser) parseVersion(f *ast.File) error {
	// Orbit file example:
	/*
		version 5
	*/

	if f.Version != nil {
		// Only allowed once.
		return p.errorf("duplicate version")
	}
	f.Version = &ast.Version{Pos: p.tk.Pos}

	// Version is an integer and must be positive.
	var err error
	f.Version.Value, err = p.expectInt()
	if err != nil {
		return err
	} else if f.Version.Value <= 0 {
		return p.errorf("version must be positive")
	}

	return nil
}

func (p *parser) parseEnum(f *ast.File) error {
	// Orbit file example:
	/*
		enum vehicle {
			car = 1
			pickup = 2
			...
		}
	*/
	e := &ast.Enum{Pos: p.tk.Pos}

	// Identifier.
	var err error
	e.Name, err = p.expectIdent()
	if err != nil {
		return err
	}

	// '{'.
	err = p.expectToken(lexer.LBRACE)
	if err != nil {
		return err
	}

	// Enum values.
	for !p.checkToken(lexer.RBRACE) {
		ev := &ast.EnumValue{}

		// Name.
		ev.Name, err = p.expectIdent()
		if err != nil {
			return err
		}
		ev.Pos = p.tk.Pos

		// '='.
		err = p.expectToken(lexer.EQUAL)
		if err != nil {
			return err
		}

		// Value.
		ev.Value, err = p.expectInt()
		if err != nil {
			return err
		}

		// Add new enum value to enum.
		e.Values = append(e.Values, ev)
	}

	f.Enums = append(f.Enums, e)
	return nil
}

func (p *parser) parseErrors(f *ast.File) error {
	// Orbit file example:
	/*
		errors {
			thisIsATest = 1
			iAmAnError = 2
		}
	*/

	// '{'.
	err := p.expectToken(lexer.LBRACE)
	if err != nil {
		return err
	}

	// Error values.
	for !p.checkToken(lexer.RBRACE) {
		e := &ast.Error{}

		// Value identifier.
		e.Name, err = p.expectIdent()
		if err != nil {
			return err
		}
		e.Pos = p.tk.Pos

		// '='.
		err = p.expectToken(lexer.EQUAL)
		if err != nil {
			return err
		}

		// ID.
		e.ID, err = p.expectInt()
		if err != nil {
			return err
		}

		// Add new error to errors.
		f.Errs = append(f.Errs, e)
	}

	return nil
}

func (p *parser) parseTypeDeclaration(f *ast.File) error {
	// Orbit file example:
	/*
		type info {
			name    string `validate:"required,min=1"`
			age     int    `validate:"required,min=1,max=155"`
			locale  string `validate:"required,len=5"`
			address string
		}
	*/
	t := &ast.Type{Pos: p.tk.Pos}

	// Identifier.
	var err error
	t.Name, err = p.expectIdent()
	if err != nil {
		return err
	}

	// '{'.
	err = p.expectToken(lexer.LBRACE)
	if err != nil {
		return err
	}

	// Type definition.
	t.Fields, err = p.expectTypeDefinition()
	if err != nil {
		return err
	}

	f.Types = append(f.Types, t)
	return nil
}

func (p *parser) parseService(f *ast.File) error {
	// Orbit file example:
	/*
		service {
			call <...>
			<...>

			stream <...>
			<...>
		}
	*/

	if f.Srvc != nil {
		// Only allowed once.
		return p.errorf("duplicate service")
	}
	f.Srvc = &ast.Service{Pos: p.tk.Pos}

	// '{'
	err := p.expectToken(lexer.LBRACE)
	if err != nil {
		return err
	}

	// Expect calls and streams.
	for !p.checkToken(lexer.RBRACE) {
		if p.checkToken(lexer.CALL) {
			c, ts, err := p.expectServiceCall()
			if err != nil {
				return err
			}

			f.Srvc.Calls = append(f.Srvc.Calls, c)
			if ts != nil {
				f.Types = append(f.Types, ts...)
			}
		} else if p.checkToken(lexer.STREAM) {
			s, ts, err := p.expectServiceStream()
			if err != nil {
				return err
			}

			f.Srvc.Streams = append(f.Srvc.Streams, s)
			if ts != nil {
				f.Types = append(f.Types, ts...)
			}
		} else {
			return p.errorf("expected service-level keyword, found %s", p.tk.Value)
		}
	}

	return nil
}
