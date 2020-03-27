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
	"io"

	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/token"
)

const (
	tkErrors  = "errors"
	tkService = "service"
	tkType    = "type"
	tkEnum    = "enum"

	tkSrvcCall   = "call"
	tkSrvcStream = "stream"
	tkSrvcUrl    = "url"

	tkEntryAsync   = "async"
	tkEntryArg     = "arg"
	tkEntryRet     = "ret"
	tkEntryTimeout = "timeout"

	tkMap = "map"
)

type Parser interface {
	// Parse the tokens read from the given reader and return the
	// AST structure received from this operation.
	Parse(token.Reader) (*ast.Tree, error)
}

// Implements the Parser interface.
type parser struct {
	tr       token.Reader
	ct       *token.Token
	prevLine int

	srvc  *ast.Service
	types []*ast.Type
	errs  []*ast.Error
	enums []*ast.Enum
}

func NewParser() Parser {
	return &parser{}
}

func (p *parser) reset(tr token.Reader) (err error) {
	// Prepare the parser and load the first token.
	p.tr = tr
	p.prevLine = 1
	p.ct, err = tr.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = nil
		}
		return
	}

	// Reset.
	p.srvc = nil
	p.types = make([]*ast.Type, 0)
	p.errs = make([]*ast.Error, 0)
	p.enums = make([]*ast.Enum, 0)

	return
}

func (p *parser) Parse(tr token.Reader) (tree *ast.Tree, err error) {
	// Reset the parser.
	err = p.reset(tr)
	if err != nil {
		return
	}

	for {
		if p.empty() {
			break
		} else if p.checkSymbol(tkService) {
			// Check for duplicate.
			if p.srvc != nil {
				err = ast.NewErr(p.line(), "duplicate service")
				return
			}

			// Expect service.
			err = p.expectService()
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkType) {
			// Expect global type.
			err = p.expectType()
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkEnum) {
			// Expect global enum.
			err = p.expectEnum()
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkErrors) {
			// Expect global errors.
			err = p.expectErrors()
			if err != nil {
				return
			}
		} else {
			err = ast.NewErr(p.line(), "unknown top-level keyword '%s'", p.value())
			return
		}
	}

	tree = &ast.Tree{
		Srvc:  p.srvc,
		Types: p.types,
		Errs:  p.errs,
		Enums: p.enums,
	}
	return
}

// Returns true, if no more token is available.
func (p *parser) empty() bool {
	return p.ct == nil
}

// Advances to the next token, discarding the current token.
// Returns io.EOF, if no more tokens are available.
func (p *parser) next() (err error) {
	if p.empty() {
		return io.EOF
	}

	// Save previous line.
	p.prevLine = p.ct.Line

	// Load next token.
	p.ct, err = p.tr.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = nil
		}
		return
	}
	return
}

// Returns the value of the current token, or the empty string,
// if none is available.
func (p *parser) value() string {
	if p.empty() {
		return ""
	}
	return p.ct.Value
}

// Returns the line of the current token, or 0, if none available.
func (p *parser) line() int {
	if p.empty() {
		return 0
	}
	return p.ct.Line
}
