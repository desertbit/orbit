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

package gen

import (
	"sort"

	"github.com/desertbit/orbit/internal/codegen/ast"
)

func (g *generator) genTypes(ts []*ast.Type, srvcs []*ast.Service, streamChanSize uint) {
	// Sort the types in alphabetical order.
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Name < ts[j].Name
	})

NextType:
	for _, t := range ts {
		// Sort its fields in alphabetical order.
		sort.Slice(t.Fields, func(i, j int) bool {
			return t.Fields[i].Name < t.Fields[j].Name
		})

		writeLn("type %s struct {", t.Name)
		for _, f := range t.Fields {
			writeLn("%s %s", f.Name, f.DataType.String())
		}
		writeLn("}")
		writeLn("")

		// Generate a chan type, if it is used in a stream as arg or ret value.
		for _, srvc := range srvcs {
			for _, s := range srvc.Streams {
				if (s.Args != nil && s.Args.Name == t.Name) || (s.Ret != nil && s.Ret.Name == t.Name) {
					g.genChanType(t.Name, false, streamChanSize)
					g.genChanType(t.Name, true, streamChanSize)
					continue NextType
				}
			}
		}
	}
}

func (g *generator) genChanType(name string, ro bool, streamChanSize uint) {
	suffix := "Write"
	if ro {
		suffix = "Read"
	}

	// Type definition.
	writeLn("//msgp:ignore %s%sChan", name, suffix)
	writeLn("type %s%sChan struct {", name, suffix)
	writeLn("closer.Closer")
	write("C ")
	if ro {
		write("<-chan ")
	} else {
		write("chan<- ")
	}
	writeLn("*%s", name)
	writeLn("c chan *%s", name)
	writeLn("mx sync.Mutex")
	writeLn("err error")
	writeLn("}")
	writeLn("")

	// Constructor.
	writeLn("func new%s%sChan(cl closer.Closer) *%s%sChan {", name, suffix, name, suffix)
	writeLn("c := &%s%sChan{Closer: cl, c: make(chan *%s, %d)}", name, suffix, name, streamChanSize)
	writeLn("c.C = c.c")
	writeLn("return c")
	writeLn("}")
	writeLn("")

	// setError method.
	writeLn("func (c *%s%sChan) setError(err error) {", name, suffix)
	writeLn("c.mx.Lock()")
	writeLn("c.err = err")
	writeLn("c.mx.Unlock()")
	writeLn("c.Close_()")
	writeLn("}")
	writeLn("")

	// Err method.
	writeLn("func (c *%s%sChan) Err() (err error) {", name, suffix)
	writeLn("c.mx.Lock()")
	writeLn("err = c.err")
	writeLn("c.mx.Unlock()")
	writeLn("return")
	writeLn("}")
	writeLn("")
}
