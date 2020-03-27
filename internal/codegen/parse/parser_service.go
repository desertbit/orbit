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
		if p.checkSymbol(token.BraceR) {
			// End of service.
			break
		} else if p.checkSymbol(tkSrvcCall) {
			// Expect call.
			c, err = p.expectServiceCall()
			if err != nil {
				return
			}

			srvc.Calls = append(srvc.Calls, c)
		} else if p.checkSymbol(tkSrvcStream) {
			// Expect stream.
			s, err = p.expectServiceStream()
			if err != nil {
				return
			}

			srvc.Streams = append(srvc.Streams, s)
		} else if p.checkSymbol(tkSrvcUrl) {
			// Check for duplicate.
			if srvc.Url != "" {
				err = ast.NewErr(p.line(), "duplicate url")
				return
			}

			// Consume ':'.
			err = p.expectSymbol(token.Colon)
			if err != nil {
				return
			}

			// Expect url.
			srvc.Url, err = p.expectServiceUrl()
			if err != nil {
				return
			}
		} else {
			err = ast.NewErr(p.line(), "unexpected symbol '%s'", p.value())
			return
		}
	}

	// Add to parser.
	p.srvc = srvc

	return
}

// Returns ast.Err.
func (p *parser) expectServiceUrl() (url string, err error) {
	if p.empty() {
		err = ast.NewErr(p.prevLine, "expected url, but is missing")
		return
	}

	url = p.value()

	// Advance to the next token.
	err = p.next()
	if err != nil {
		return
	}
	return
}

// Returns ast.Err.
func (p *parser) expectServiceCall() (c *ast.Call, err error) {
	c = &ast.Call{}

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
		} else if p.checkSymbol(tkEntryArg) {
			// Check for duplicate.
			if c.Arg != nil {
				err = ast.NewErr(p.line(), "duplicate arg")
				return
			}

			// Consume ':'.
			err = p.expectSymbol(token.Colon)
			if err != nil {
				return
			}

			// Parse args.
			c.Arg, c.ArgValTag, err = p.expectServiceEntryType(c.Name + "Arg")
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkEntryRet) {
			// Check for duplicate.
			if c.Ret != nil {
				err = ast.NewErr(p.line(), "duplicate ret")
				return
			}

			// Consume ':'.
			err = p.expectSymbol(token.Colon)
			if err != nil {
				return
			}

			// Parse ret.
			c.Ret, c.RetValTag, err = p.expectServiceEntryType(c.Name + "Ret")
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkEntryAsync) {
			// Check for duplicate.
			if c.Async {
				err = ast.NewErr(p.line(), "duplicate async")
				return
			}

			c.Async = true
		} else if p.checkSymbol(tkEntryTimeout) {
			// Check for duplicate.
			if c.Timeout != nil {
				err = ast.NewErr(p.line(), "duplicate timeout")
				return
			}

			// Consume ':'.
			err = p.expectSymbol(token.Colon)
			if err != nil {
				return
			}

			// Parse timeout.
			c.Timeout, err = p.expectTimeDuration()
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
func (p *parser) expectServiceStream() (s *ast.Stream, err error) {
	s = &ast.Stream{}

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
		} else if p.checkSymbol(tkEntryArg) {
			// Check for duplicate.
			if s.Arg != nil {
				err = errors.New("duplicate arg")
				return
			}

			// Consume ':'.
			err = p.expectSymbol(token.Colon)
			if err != nil {
				return
			}

			// Parse args.
			s.Arg, s.ArgValTag, err = p.expectServiceEntryType(s.Name + "Arg")
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkEntryRet) {
			// Check for duplicate.
			if s.Ret != nil {
				err = errors.New("duplicate ret")
				return
			}

			// Consume ':'.
			err = p.expectSymbol(token.Colon)
			if err != nil {
				return
			}

			// Parse ret.
			s.Ret, s.RetValTag, err = p.expectServiceEntryType(s.Name + "Ret")
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
func (p *parser) expectServiceEntryType(name string) (dt ast.DataType, valTag string, err error) {
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

	// The entry has a normal data type.
	dt, err = p.expectDataType()
	if err != nil {
		return
	}

	// Check for a validation tag definition.
	if p.checkSymbol(token.SingQuote) {
		valTag, err = p.expectValTag()
		if err != nil {
			return
		}
	}

	return
}
