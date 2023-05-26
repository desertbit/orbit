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
	"fmt"
	"math"
	"strconv"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/lexer"
)

type stateFn func(p *parser, f *ast.File) stateFn

type parser struct {
	lx lexer.Lexer

	tk       lexer.Token // The current token from the lexer.
	unreadTk lexer.Token // A previous current token that has been unread.
}

// Parse parses the output of the lexer until an EOF or an error is encountered.
// If error is nil, the returned ast.File contains the parsed input.
func Parse(lx lexer.Lexer) (*ast.File, error) {
	// Create the parser.
	p := &parser{lx: lx}

	// Run the parser.
	return p.run()
}

func (p *parser) expectIdent() (string, error) {
	err := p.next()
	if err != nil {
		return "", err
	} else if p.tk.Type != lexer.IDENT {
		return "", p.errorf("expected identifier, got %s", p.tk.Value)
	}

	return p.tk.Value, nil
}

func (p *parser) expectInt() (int, error) {
	err := p.next()
	if err != nil {
		return 0, err
	} else if p.tk.Type != lexer.INT {
		return 0, p.errorf("expected integer, got %s", p.tk.Value)
	}

	i, err := strconv.Atoi(p.tk.Value)
	if err != nil {
		return 0, p.errorf("expected integer, but failed to parse %s, %v", p.tk.Value, err)
	}

	return i, nil
}

func (p *parser) expectDuration() (time.Duration, error) {
	err := p.next()
	if err != nil {
		return 0, err
	} else if p.tk.Type != lexer.IDENT {
		return 0, p.errorf("expected time duration identifier, got %s", p.tk.Value)
	}

	dur, err := time.ParseDuration(p.tk.Value)
	if err != nil {
		return 0, p.errorf("expected time duration, but failed to parse %s, %v", p.tk.Value, err)
	}

	return dur, nil
}

// Can return -1 as special value.
func (p *parser) expectByteSize() (int64, error) {
	err := p.next()
	if err != nil {
		return 0, err
	} else if p.tk.Type == lexer.INT {
		// Check for special '-1' value.
		if p.tk.Value != "-1" {
			return 0, p.errorf("expected byte size, got integer %s", p.tk.Type)
		}

		return -1, nil
	} else if p.tk.Type != lexer.IDENT {
		return 0, p.errorf("expected byte size identifier, got %s", p.tk.Value)
	}

	size, err := bytefmt.ToBytes(p.tk.Value)
	if err != nil {
		return 0, p.errorf("expected byte size, but failed to parse %s, %v", p.tk.Value, err)
	}

	// Check, that the size is not larger than uint32.
	if size > uint64(math.MaxUint32) {
		return 0, p.errorf("byte size too large, max is %d", math.MaxUint32)
	}

	return int64(size), nil
}

func (p *parser) expectToken(dl lexer.TokenType) error {
	err := p.next()
	if err != nil || p.tk.Type != dl {
		return p.errorf("expected symbol %s, got %s", dl.String(), p.tk.Value)
	}

	return nil
}

func (p *parser) checkToken(dl lexer.TokenType) bool {
	if p.next() != nil {
		return false
	} else if p.tk.Type != dl {
		p.backup()
		return false
	}

	return true
}

func (p *parser) next() error {
	// Check, if a token has been unread, which must be consumed first.
	if p.unreadTk != (lexer.Token{}) {
		p.tk = p.unreadTk
		p.unreadTk = lexer.Token{}
		return nil
	}

	p.tk = p.lx.Next()
	if p.tk.Type == lexer.ILLEGAL {
		// Backup the illegal token so every p.next call returns it again.
		p.backup()
		return p.errorf(p.tk.Value)
	}

	return nil
}

func (p *parser) backup() {
	p.unreadTk = p.tk
}

func (p *parser) errorf(format string, args ...interface{}) error {
	format += " (pos %d:%d)"
	args = append(args, p.tk.Pos.Line, p.tk.Pos.Column)
	return fmt.Errorf(format, args...)
}
