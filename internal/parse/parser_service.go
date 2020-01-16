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
func (p *parser) expectService() (err error) {
	defer func() {
		var pErr *Err
		if err != nil && !errors.As(err, &pErr) {
			err = &Err{msg: err.Error(), line: p.ct.line}
		}
	}()

	// Expect name.
	name, err := p.expectName()
	if err != nil {
		return
	}

	// Check for duplicates.
	if _, ok := p.srvcs[name]; ok {
		return fmt.Errorf("service '%s' declared twice", name)
	}

	// Expecting "{"
	err = p.expectSymbol(tkBraceL)
	if err != nil {
		return
	}

	// Init the struct map.
	p.srvcsStructs[name] = make([]*StructType, 0)

	var entries []Entry

	for {
		if p.checkSymbol(tkBraceR) {
			// End of service.
			break
		} else if p.checkSymbol(tkErrors) {
			// Expect errors.
			err = p.expectErrors(name)
			if err != nil {
				return
			}
		} else if p.checkSymbol(tkType) {
			// Expect type and prepend the service name to the type.
			var sts []*StructType
			sts, err = p.expectType(name, "")
			if err != nil {
				return
			}

			p.srvcsStructs[name] = append(p.srvcsStructs[name], sts...)
		} else {
			// Expect entry.
			var (
				e   Entry
				sts []*StructType
			)
			e, sts, err = p.expectServiceEntry(name)
			if err != nil {
				return
			}

			// Check for duplicate entries.
			for _, en := range entries {
				if en.NamePrv() == e.NamePrv() {
					err = fmt.Errorf("entry '%s' declared twice in service '%s'", e.NamePrv(), name)
					return
				}
			}

			entries = append(entries, e)
			p.srvcsStructs[name] = append(p.srvcsStructs[name], sts...)
		}
	}

	// There may have been some types defined inside the service with the same name as outer types.
	// In order to distinguish between them, we must now check, if any entry references a type,
	// whose name, prepended with the service's name, is equal to a type defined inside this service.
	// The same goes for any type fields, where a struct references a service local type.
	// FIXME: this whole thing is way too complicated right now. Maybe there is a better solution.
	for tn := range p.types {
		for _, s := range p.srvcsStructs[name] {
			if name+s.Name == tn {
				s.Name = name + s.Name
			}
		}
	}

	// Save service and free resources.
	p.srvcs[name] = &Service{Name: name, Entries: entries}
	delete(p.srvcsStructs, name)

	return
}

// Returns Err.
func (p *parser) expectServiceEntry(srvcName string) (e Entry, sts []*StructType, err error) {
	defer func() {
		var pErr *Err
		if err != nil && !errors.As(err, &pErr) {
			err = &Err{msg: err.Error(), line: p.ct.line}
		}
	}()

	// Check for async.
	async := p.checkSymbol(tkEntryAsync)

	// Expect the entry type.
	if !p.next() {
		err = errors.New("expected entry type (rev)call or (rev)stream, but is missing")
		return
	}
	t := p.ct.value

	// Validate type.
	if t != tkEntryCall && t != tkEntryRevCall && t != tkEntryStream && t != tkEntryRevStream {
		err = fmt.Errorf("expected entry type (rev)call or (rev)stream, but got '%s'", p.ct.value)
		return
	} else if async && (t == tkEntryStream || t == tkEntryRevStream) {
		err = errors.New("async only allowed with calls and revcalls")
		return
	}

	// Expect name.
	name, err := p.expectName()
	if err != nil {
		return
	}

	// Expect '{'.
	err = p.expectSymbol(tkBraceL)
	if err != nil {
		return
	}

	var args, ret *StructType

	// Check for arguments.
	if p.checkSymbol(tkEntryArgs) {
		var asts []*StructType
		args, asts, err = p.expectServiceEntryType(srvcName, name+"Args")
		if err != nil {
			return
		}

		sts = append(sts, asts...)
	}

	// Check for returns.
	if p.checkSymbol(tkEntryRet) {
		var rsts []*StructType
		ret, rsts, err = p.expectServiceEntryType(srvcName, name+"Ret")
		if err != nil {
			return
		}

		sts = append(sts, rsts...)
	}

	// Expect '}'.
	err = p.expectSymbol(tkBraceR)
	if err != nil {
		return
	}

	// Create entry based on type.
	if t == tkEntryCall || t == tkEntryRevCall {
		e = &Call{name: name, rev: t == tkEntryRevCall, Async: async, args: args, ret: ret}
	} else {
		e = &Stream{name: name, rev: t == tkEntryRevStream, args: args, ret: ret}
	}
	return
}

// Returns Err.
func (p *parser) expectServiceEntryType(srvcName, name string) (st *StructType, sts []*StructType, err error) {
	defer func() {
		var pErr *Err
		if err != nil && !errors.As(err, &pErr) {
			err = &Err{msg: err.Error(), line: p.ct.line}
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
