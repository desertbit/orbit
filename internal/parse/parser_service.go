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
func (p *parser) expectService() (srvc *Service, err error) {
	defer func() {
		var pErr *Err
		if err != nil && !errors.As(err, &pErr) {
			err = &Err{msg: err.Error(), line: p.lt.line}
		}
	}()

	// Expect name.
	name, err := p.expectName()
	if err != nil {
		return
	}

	srvc = &Service{Name: name, line: p.lt.line}

	// Expecting "{"
	err = p.expectSymbol(tkBraceL)
	if err != nil {
		return
	}

	var (
		errs []*Error
		t    *Type
		c    *Call
		s    *Stream

		srvcsStructs = make([]*StructType, 0)
	)

	// Parse fields.
	for {
		var (
			sts []*StructType

			revCall   = p.checkSymbol(tkEntryRevCall)
			revStream = p.checkSymbol(tkEntryRevStream)
		)

		if p.checkSymbol(tkBraceR) {
			// End of service.
			break
		} else if p.checkSymbol(tkErrors) {
			// Expect errors and prepend the service name to them.
			errs, err = p.expectErrors(name)
			if err != nil {
				return
			}

			srvc.Errors = append(srvc.Errors, errs...)
		} else if p.checkSymbol(tkType) {
			// Expect type and prepend the service name to it.
			t, sts, err = p.expectType(name, "")
			if err != nil {
				return
			}

			srvc.Types = append(srvc.Types, t)
		} else if revCall || p.checkSymbol(tkEntryCall) {
			// Expect call and prepend the service name to it.
			c, sts, err = p.expectServiceCall(name, revCall)
			if err != nil {
				return
			}

			srvc.Calls = append(srvc.Calls, c)
		} else if revStream || p.checkSymbol(tkEntryStream) {
			// Expect stream and prepend the service name to it.
			s, sts, err = p.expectServiceStream(name, revStream)
			if err != nil {
				return
			}

			srvc.Streams = append(srvc.Streams, s)
		} else {
			err = fmt.Errorf("unexpected symbol '%s'", p.ct.value)
			return
		}

		if len(sts) > 0 {
			srvcsStructs = append(srvcsStructs, sts...)
		}
	}

	// There may have been some types defined inside the service with the same name as outer types.
	// In order to distinguish between them, we must now check, if any entry references a type,
	// whose name, prepended with the service's name, is equal to a type defined inside this service.
	// The same goes for any type fields, where a struct references a service local type.
	// FIXME: this whole thing is way too complicated right now. Maybe there is a better solution.
	for tn := range p.types {
		for _, s := range srvcsStructs {
			if name+s.Name == tn {
				s.Name = name + s.Name
			}
		}
	}

	return
}

// Returns Err.
func (p *parser) expectServiceCall(srvcName string, rev bool) (c *Call, sts []*StructType, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parsing call of service '%s': %w", err)

			var pErr *Err
			if !errors.As(err, &pErr) {
				err = &Err{msg: err.Error(), line: p.lt.line}
			}
		}
	}()

	// Expect name.
	name, err := p.expectName()
	if err != nil {
		return
	}

	c = &Call{Name: name, Rev: rev, line: p.lt.line}

	// Expect '{'.
	err = p.expectSymbol(tkBraceL)
	if err != nil {
		return
	}

	// Parse fields.
	for {
		// Check for end.
		if p.checkSymbol(tkBraceR) {
			return
		} else if p.checkSymbol(tkEntryArgs) {
			// Check for duplicate.
			if c.Args != nil {
				err = errors.New("double args")
				return
			}

			// Parse args.
			var asts []*StructType
			c.Args, asts, err = p.expectServiceEntryType(srvcName, name+"Args")
			if err != nil {
				return
			}

			sts = append(sts, asts...)
		} else if p.checkSymbol(tkEntryRet) {
			// Check for duplicate.
			if c.Ret != nil {
				err = errors.New("double ret")
				return
			}

			// Parse ret.
			var rsts []*StructType
			c.Ret, rsts, err = p.expectServiceEntryType(srvcName, name+"Ret")
			if err != nil {
				return
			}

			sts = append(sts, rsts...)
		} else if p.checkSymbol(tkEntryAsync) {
			// Check for duplicate.
			if c.Async {
				err = errors.New("double async")
			}
		} else {
			err = fmt.Errorf("unexpected symbol '%s'", p.ct.value)
			return
		}
	}
}

// Returns Err.
func (p *parser) expectServiceStream(srvcName string, rev bool) (s *Stream, sts []*StructType, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parsing stream of service '%s': %w", err)

			var pErr *Err
			if !errors.As(err, &pErr) {
				err = &Err{msg: err.Error(), line: p.lt.line}
			}
		}
	}()

	// Expect name.
	name, err := p.expectName()
	if err != nil {
		return
	}

	s = &Stream{Name: name, Rev: rev, line: p.lt.line}

	// Expect '{'.
	err = p.expectSymbol(tkBraceL)
	if err != nil {
		return
	}

	// Parse fields.
	for {
		// Check for end.
		if p.checkSymbol(tkBraceR) {
			return
		} else if p.checkSymbol(tkEntryArgs) {
			// Check for duplicate.
			if s.Args != nil {
				err = errors.New("double args")
				return
			}

			// Parse args.
			var asts []*StructType
			s.Args, asts, err = p.expectServiceEntryType(srvcName, name+"Args")
			if err != nil {
				return
			}

			sts = append(sts, asts...)
		} else if p.checkSymbol(tkEntryRet) {
			// Check for duplicate.
			if s.Ret != nil {
				err = errors.New("double ret")
				return
			}

			// Parse ret.
			var rsts []*StructType
			s.Ret, rsts, err = p.expectServiceEntryType(srvcName, name+"Ret")
			if err != nil {
				return
			}

			sts = append(sts, rsts...)
		} else {
			err = fmt.Errorf("unexpected symbol '%s'", p.ct.value)
			return
		}
	}
}

// Returns Err.
func (p *parser) expectServiceEntryType(srvcName, name string) (st *StructType, sts []*StructType, err error) {
	defer func() {
		var pErr *Err
		if err != nil && !errors.As(err, &pErr) {
			err = &Err{msg: err.Error(), line: p.lastLine}
		}
	}()

	// Check for an inline type definition.
	if p.peekSymbol(tkBraceL) {
		structName := srvcName + name

		sts, err = p.expectType(srvcName, name)
		if err != nil {
			err = fmt.Errorf("parsing service entry type '%s': %w", structName, err)
			return
		}

		// The struct type is a reference to the inline type.
		st = &StructType{Name: structName}
		sts = append(sts, st)
		return
	}

	// The entry type must be a struct type.
	st, err = p.expectStructType()
	if err != nil {
		return
	}

	sts = append(sts, st)
	return
}
