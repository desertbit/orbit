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
	"github.com/desertbit/orbit/internal/utils"
)

func (p *parser) expectServiceCall() (*ast.Call, []*ast.Type, error) {
	// Orbit file example:
	/*
		call test {
			async
			arg: {
				s string
			}
			ret: {
				name string `validate:"required,min=1"`
				ts time
			}
			timeout: 500ms
			maxRetSize: 10K
		}
	*/
	var (
		c = &ast.Call{}

		ts  []*ast.Type
		err error
	)

	// Identifier.
	c.Name, err = p.expectIdent()
	if err != nil {
		return nil, nil, err
	}
	// Call name must be uppercase first.
	c.Name = utils.FirstUpper(c.Name)

	// '{'.
	err = p.expectToken(lexer.LBRACE)
	if err != nil {
		return nil, nil, err
	}

	// Parse fields.
	for !p.checkToken(lexer.RBRACE) {
		if p.checkToken(lexer.ARG) {
			// Check for duplicate.
			if c.Arg != nil {
				return nil, nil, p.errorf("duplicate arg")
			}

			// ':'.
			err = p.expectToken(lexer.COLON)
			if err != nil {
				return nil, nil, err
			}

			// Parse args.
			var t *ast.Type
			c.Arg, t, err = p.expectServiceEntryType(c.Name + "Arg")
			if err != nil {
				return nil, nil, err
			}

			// Add inline type, if one was found.
			if t != nil {
				ts = append(ts, t)
			}
		} else if p.checkToken(lexer.RET) {
			// Check for duplicate.
			if c.Ret != nil {
				return nil, nil, p.errorf("duplicate ret")
			}

			// Consume ':'.
			err = p.expectToken(lexer.COLON)
			if err != nil {
				return nil, nil, err
			}

			// Parse ret.
			var t *ast.Type
			c.Ret, t, err = p.expectServiceEntryType(c.Name + "Ret")
			if err != nil {
				return nil, nil, err
			}

			// Add inline type, if one was found.
			if t != nil {
				ts = append(ts, t)
			}
		} else if p.checkToken(lexer.ASYNC) {
			// Check for duplicate.
			if c.Async {
				return nil, nil, p.errorf("duplicate async")
			}

			c.Async = true
		} else if p.checkToken(lexer.TIMEOUT) {
			// Check for duplicate.
			if c.Timeout != nil {
				return nil, nil, p.errorf("duplicate timeout")
			}

			// Consume ':'.
			err = p.expectToken(lexer.COLON)
			if err != nil {
				return nil, nil, err
			}

			// Parse duration.
			d, err := p.expectDuration()
			if err != nil {
				return nil, nil, err
			}
			c.Timeout = &d
		} else if p.checkToken(lexer.MAXARGSIZE) {
			// Check for duplicate.
			if c.MaxArgSize != nil {
				return nil, nil, p.errorf("duplicate maxArgSize")
			}

			// ':'.
			err = p.expectToken(lexer.COLON)
			if err != nil {
				return nil, nil, err
			}

			// Expect a byte size.
			size, err := p.expectByteSize()
			if err != nil {
				return nil, nil, err
			}
			c.MaxArgSize = &size
		} else if p.checkToken(lexer.MAXRETSIZE) {
			// Check for duplicate.
			if c.MaxRetSize != nil {
				return nil, nil, p.errorf("duplicate maxArgSize")
			}

			// ':'.
			err = p.expectToken(lexer.COLON)
			if err != nil {
				return nil, nil, err
			}

			// Expect a data size entry.
			size, err := p.expectByteSize()
			if err != nil {
				return nil, nil, err
			}
			c.MaxRetSize = &size
		} else {
			return nil, nil, p.errorf("unexpected token in service call '%s'", p.tk.Value)
		}
	}

	return c, ts, nil
}

func (p *parser) expectServiceStream() (*ast.Stream, []*ast.Type, error) {
	// Orbit file example:
	/*
		stream bidirectional {
			maxArgSize: 100K
			arg: {
				question string
			}
			ret: {
				answer string
			}
		}
	*/
	var (
		s = &ast.Stream{}

		ts  []*ast.Type
		err error
	)

	// Identifier.
	s.Name, err = p.expectIdent()
	if err != nil {
		return nil, nil, err
	}
	// Stream name must be uppercase first.
	s.Name = utils.FirstUpper(s.Name)

	// '{'.
	err = p.expectToken(lexer.LBRACE)
	if err != nil {
		return nil, nil, err
	}

	// Parse fields.
	for !p.checkToken(lexer.RBRACE) {
		if p.checkToken(lexer.ARG) {
			// Check for duplicate.
			if s.Arg != nil {
				return nil, nil, p.errorf("duplicate arg")
			}

			// ':'.
			err = p.expectToken(lexer.COLON)
			if err != nil {
				return nil, nil, err
			}

			// Parse args.
			var t *ast.Type
			s.Arg, t, err = p.expectServiceEntryType(s.Name + "Arg")
			if err != nil {
				return nil, nil, err
			}

			// Add inline type, if one was found.
			if t != nil {
				ts = append(ts, t)
			}
		} else if p.checkToken(lexer.RET) {
			// Check for duplicate.
			if s.Ret != nil {
				return nil, nil, p.errorf("duplicate ret")
			}

			// Consume ':'.
			err = p.expectToken(lexer.COLON)
			if err != nil {
				return nil, nil, err
			}

			// Parse ret.
			var t *ast.Type
			s.Ret, t, err = p.expectServiceEntryType(s.Name + "Ret")
			if err != nil {
				return nil, nil, err
			}

			// Add inline type, if one was found.
			if t != nil {
				ts = append(ts, t)
			}
		} else if p.checkToken(lexer.MAXARGSIZE) {
			// Check for duplicate.
			if s.MaxArgSize != nil {
				return nil, nil, p.errorf("duplicate maxArgSize")
			}

			// ':'.
			err = p.expectToken(lexer.COLON)
			if err != nil {
				return nil, nil, err
			}

			// Expect a byte size.
			size, err := p.expectByteSize()
			if err != nil {
				return nil, nil, err
			}
			s.MaxArgSize = &size
		} else if p.checkToken(lexer.MAXRETSIZE) {
			// Check for duplicate.
			if s.MaxRetSize != nil {
				return nil, nil, p.errorf("duplicate maxArgSize")
			}

			// ':'.
			err = p.expectToken(lexer.COLON)
			if err != nil {
				return nil, nil, err
			}

			// Expect a data size entry.
			size, err := p.expectByteSize()
			if err != nil {
				return nil, nil, err
			}
			s.MaxRetSize = &size
		} else {
			return nil, nil, p.errorf("unexpected token in service call '%s'", p.tk.Value)
		}
	}

	return s, ts, nil
}

func (p *parser) expectServiceEntryType(name string) (ast.DataType, *ast.Type, error) {
	// Check for an inline type definition.
	if p.checkToken(lexer.LBRACE) {
		// The struct type is a reference to the inline type.
		dt := &ast.StructType{NamePrv: name}

		// Expect the inline type definition and add it to the global types.
		t := &ast.Type{Name: name}

		var err error
		t.Fields, err = p.expectTypeDefinition()
		if err != nil {
			return nil, nil, err
		}

		return dt, t, nil
	}

	// The entry has an any type.
	dt, err := p.expectAnyType()
	return dt, nil, err
}
