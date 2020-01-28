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

	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/token"
)

// Returns ast.Err.
func (p *parser) expectService() (err error) {
	srvc := &ast.Service{}

	// Expect name.
	srvc.Name, srvc.Line, err = p.expectName()
	if err != nil {
		return
	}

	// Expect '{'.
	err = p.expectSymbol(token.BraceL)
	if err != nil {
		return
	}

	var (
		c *ast.Call
		s *ast.Stream
	)

	// Parse fields.
	for {
		var (
			revCall   = p.checkSymbol(tkEntryRevCall)
			revStream = p.checkSymbol(tkEntryRevStream)
		)

		if p.checkSymbol(token.BraceR) {
			// End of service.
			break
		} else if revCall || p.checkSymbol(tkEntryCall) {
			// Expect call.
			c, err = p.expectServiceCall(revCall)
			if err != nil {
				return
			}

			srvc.Calls = append(srvc.Calls, c)
		} else if revStream || p.checkSymbol(tkEntryStream) {
			// Expect stream.
			s, err = p.expectServiceStream(revStream)
			if err != nil {
				return
			}

			srvc.Streams = append(srvc.Streams, s)
		} else {
			err = ast.NewErr(p.line(), "unexpected symbol '%s'", p.value())
			return
		}
	}

	// Add to services.
	p.srvcs = append(p.srvcs, srvc)

	return
}

// Returns ast.Err.
func (p *parser) expectServiceCall(rev bool) (c *ast.Call, err error) {
	c = &ast.Call{Rev: rev}

	// Expect name.
	c.Name, c.Line, err = p.expectName()
	if err != nil {
		return
	}

	// Expect '{'.
	err = p.expectSymbol(token.BraceL)
	if err != nil {
		return
	}

	// Parse fields.
	for {
		// Check for end.
		if p.checkSymbol(token.BraceR) {
			return
		} else if p.checkSymbol(tkEntryArgs) {
			// Check for duplicate.
			if c.Args != nil {
				err = errors.New("double args")
				return
			}

			// Consume ':', if present.
			_ = p.checkSymbol(token.Colon)

			// Parse args.
			c.Args, err = p.expectServiceEntryType(c.Name + "Args")
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkEntryRet) {
			// Check for duplicate.
			if c.Ret != nil {
				err = errors.New("double ret")
				return
			}

			// Consume ':', if present.
			_ = p.checkSymbol(token.Colon)

			// Parse ret.
			c.Ret, err = p.expectServiceEntryType(c.Name + "Ret")
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkEntryAsync) {
			// Check for duplicate.
			if c.Async {
				err = errors.New("double async")
				return
			}

			c.Async = true
		} else {
			err = ast.NewErr(p.line(), "unexpected symbol '%s'", p.value())
			return
		}
	}
}

// Returns ast.Err.
func (p *parser) expectServiceStream(rev bool) (s *ast.Stream, err error) {
	s = &ast.Stream{Rev: rev}

	// Expect name.
	s.Name, s.Line, err = p.expectName()
	if err != nil {
		return
	}

	// Expect '{'.
	err = p.expectSymbol(token.BraceL)
	if err != nil {
		return
	}

	// Parse fields.
	for {
		// Check for end.
		if p.checkSymbol(token.BraceR) {
			return
		} else if p.checkSymbol(tkEntryArgs) {
			// Check for duplicate.
			if s.Args != nil {
				err = errors.New("double args")
				return
			}

			// Consume ':', if present.
			_ = p.checkSymbol(token.Colon)

			// Parse args.
			s.Args, err = p.expectServiceEntryType(s.Name + "Args")
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkEntryRet) {
			// Check for duplicate.
			if s.Ret != nil {
				err = errors.New("double ret")
				return
			}

			// Consume ':', if present.
			_ = p.checkSymbol(token.Colon)

			// Parse ret.
			s.Ret, err = p.expectServiceEntryType(s.Name + "Ret")
			if err != nil {
				return
			}
		} else {
			err = ast.NewErr(p.line(), "unexpected symbol '%s'", p.value())
			return
		}
	}
}

// Returns ast.Err.
func (p *parser) expectServiceEntryType(name string) (dt ast.DataType, err error) {
	// Check for an inline type definition.
	if p.peekSymbol(token.BraceL) {
		// The struct type is a reference to the inline type.
		line := p.line()
		dt = &ast.StructType{NamePrv: name, Line: line}

		// Expect the inline type.
		err = p.expectInlineType(name, line)
		if err != nil {
			return
		}

		return
	}

	// The entry has is a normal data type.
	dt, err = p.expectDataType()
	if err != nil {
		return
	}

	return
}
