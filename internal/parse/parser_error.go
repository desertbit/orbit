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
	"strings"
)

// Returns Err.
func (p *parser) expectErrors() (err error) {
	defer func() {
		var pErr *Err
		if err != nil && !errors.As(err, &pErr) {
			err = &Err{msg: err.Error(), line: p.ct.line}
		}
	}()

	// Expect "{".
	err = p.expectSymbol(tkBraceL)
	if err != nil {
		return
	}

	for {
		// Check for end of service.
		if p.checkSymbol(tkBraceR) {
			return
		}

		// Expect name and ensure CamelCase.
		var name string
		name, err = p.expectName()
		if err != nil {
			return
		}
		name = strings.Title(name)

		// Check for duplicate name.
		if _, ok := p.errors[name]; ok {
			return fmt.Errorf("errors '%s' declared twice", name)
		}

		// Check, if error has already been defined.
		if _, ok := p.errors[name]; ok {
			return fmt.Errorf("error '%s' declared twice", name)
		}

		// Expect "=".
		err = p.expectSymbol(tkEqual)
		if err != nil {
			return
		}

		// Expect identifier.
		var id int
		id, err = p.expectInt()
		if err != nil {
			return
		}

		// Check for duplicate identifier.
		for _, e := range p.errors {
			if e.ID == id {
				return fmt.Errorf("errors '%s' and '%s' share same identifier", name, e.Name)
			}
		}

		// Save error.
		p.errors[name] = &Error{Name: name, ID: id}
	}
}
