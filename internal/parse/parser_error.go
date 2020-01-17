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
	"fmt"
)

// Returns Err.
func (p *parser) expectErrors(prefix string) (errs []*Error, err error) {
	// Expect "{".
	err = p.expectSymbol(tkBraceL)
	if err != nil {
		return
	}

	for {
		// Check for end of errors.
		if p.checkSymbol(tkBraceR) {
			return
		}

		// Expect name and prepend it with the service name.
		var name string
		name, err = p.expectName()
		if err != nil {
			return
		}
		name = prefix + name

		// Create error.
		e := &Error{Name: name, line: p.lt.line}

		// Expect "=".
		err = p.expectSymbol(tkEqual)
		if err != nil {
			return
		}

		// Expect identifier.
		e.ID, err = p.expectInt()
		if err != nil {
			return
		}

		// Validate.
		err = p.validateError(e)
		if err != nil {
			return
		}

		errs = append(errs, e)
	}
}

// Returns Err.
func (p *parser) validateError(e *Error) (err error) {
	defer func() {
		if err != nil {
			err = &Err{msg: err.Error(), line: e.line}
		}
	}()

	// Check for duplicate names.
	if _, ok := p.errs[e.Name]; ok {
		return fmt.Errorf("error '%s' declared twice", e.Name)
	}

	// Error ids must be greater than 0.
	if e.ID <= 0 {
		return errors.New("error ids must be greater than 0")
	}

	// Check for duplicate identifiers with the global errors.
	for _, ge := range p.errs {
		if e.ID == ge.ID {
			return fmt.Errorf("errors '%s' and '%s' share same identifier", e.Name, ge.Name)
		}
	}

	return
}
