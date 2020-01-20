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
			ts  []*Type

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
			// Expect call.
			c, ts, sts, err = p.expectServiceCall(name, revCall)
			if err != nil {
				return
			}

			// Check for duplicates.
			for _, c2 := range srvc.Calls {
				if c.Name == c2.Name {
					err = &Err{
						msg:  fmt.Sprintf("call '%s' declared twice in service '%s'", c.NamePrv(), name),
						line: c.line,
					}
					return
				}
			}

			srvc.Calls = append(srvc.Calls, c)
		} else if revStream || p.checkSymbol(tkEntryStream) {
			// Expect stream.
			s, ts, sts, err = p.expectServiceStream(name, revStream)
			if err != nil {
				return
			}

			// Check for duplicates.
			for _, s2 := range srvc.Streams {
				if s.Name == s2.Name {
					err = &Err{
						msg:  fmt.Sprintf("call '%s' declared twice in service '%s'", s.NamePrv(), name),
						line: s.line,
					}
					return
				}
			}

			srvc.Streams = append(srvc.Streams, s)
		} else {
			err = &Err{
				msg:  fmt.Sprintf("unexpected symbol '%s'", p.ct.value),
				line: p.ct.line,
			}
			return
		}

		if len(sts) > 0 {
			srvcsStructs = append(srvcsStructs, sts...)
		}
		if len(ts) > 0 {
			srvc.Types = append(srvc.Types, ts...)
		}
	}

	// Validate.
	err = p.validateService(srvc)
	if err != nil {
		return
	}

	// There may have been some types defined inside the service with the same name as outer types.
	// In order to distinguish between them, we must now check, if any entry references a type,
	// whose name, prepended with the service's name, is equal to a type defined inside this service.
	// The same goes for any type fields, where a struct references a service local type.
	// FIXME: Maybe there is a better solution?
	for _, tn := range srvc.Types {
		for _, s := range srvcsStructs {
			if name+s.Name == tn.Name {
				s.Name = name + s.Name
			}
		}
	}

	return
}

// Returns Err.
func (p *parser) validateService(srvc *Service) (err error) {
	defer func() {
		if err != nil {
			err = &Err{msg: err.Error(), line: srvc.line}
		}
	}()

	// Check for duplicate service name.
	if _, ok := p.srvcs[srvc.Name]; ok {
		return fmt.Errorf("service '%s' declared twice", srvc.Name)
	}

	return
}

// Returns Err.
func (p *parser) expectServiceCall(srvcName string, rev bool) (
	c *Call,
	ts []*Type,
	sts []*StructType,
	err error,
) {
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
		var (
			sts2 []*StructType
			t    *Type
		)

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
			c.Args, t, sts2, err = p.expectServiceEntryType(srvcName, name+"Args")
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkEntryRet) {
			// Check for duplicate.
			if c.Ret != nil {
				err = errors.New("double ret")
				return
			}

			// Parse ret.
			c.Ret, t, sts2, err = p.expectServiceEntryType(srvcName, name+"Ret")
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkEntryAsync) {
			// Check for duplicate.
			if c.Async {
				err = errors.New("double async")
			}
		} else {
			err = &Err{
				msg:  fmt.Sprintf("unexpected symbol '%s'", p.ct.value),
				line: p.ct.line,
			}
			return
		}

		sts = append(sts, sts2...)

		if t != nil {
			ts = append(ts, t)
		}
	}
}

// Returns Err.
func (p *parser) expectServiceStream(srvcName string, rev bool) (
	s *Stream,
	ts []*Type,
	sts []*StructType,
	err error,
) {
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
		var (
			sts2 []*StructType
			t    *Type
		)

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
			s.Args, t, sts2, err = p.expectServiceEntryType(srvcName, name+"Args")
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkEntryRet) {
			// Check for duplicate.
			if s.Ret != nil {
				err = errors.New("double ret")
				return
			}

			// Parse ret.
			s.Ret, t, sts2, err = p.expectServiceEntryType(srvcName, name+"Ret")
			if err != nil {
				return
			}
		} else {
			err = &Err{
				msg:  fmt.Sprintf("unexpected symbol '%s'", p.ct.value),
				line: p.ct.line,
			}
			return
		}

		sts = append(sts, sts2...)

		if t != nil {
			ts = append(ts, t)
		}
	}
}

// Returns Err.
func (p *parser) expectServiceEntryType(srvcName, name string) (
	st *StructType,
	t *Type,
	sts []*StructType,
	err error,
) {
	// Check for an inline type definition.
	if p.peekSymbol(tkBraceL) {
		structName := srvcName + name

		t, sts, err = p.expectType(srvcName, name)
		if err != nil {
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
