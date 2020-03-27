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
	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/token"
)

// Returns ast.Err.
func (p *parser) expectErrors() (err error) {
	// Expect "{".
	err = p.expectSymbol(token.BraceL)
	if err != nil {
		return
	}

	for {
		// Check for end of errors.
		if p.checkSymbol(token.BraceR) {
			return
		}

		e := &ast.Error{}

		// Expect name.
		e.Name, e.Line, err = p.expectName()
		if err != nil {
			return
		}

		// Expect "=".
		err = p.expectSymbol(token.Equal)
		if err != nil {
			return
		}

		// Expect identifier.
		e.ID, err = p.expectInt()
		if err != nil {
			return
		}

		// Add to errors.
		p.errs = append(p.errs, e)
	}
}
