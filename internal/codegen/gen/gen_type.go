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

func (g *generator) genTypes(ts []*ast.Type, srvcs []*ast.Service) {
	// Sort the types in alphabetical order.
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Name < ts[j].Name
	})

	for _, t := range ts {
		// Sort its fields in alphabetical order.
		sort.Slice(t.Fields, func(i, j int) bool {
			return t.Fields[i].Name < t.Fields[j].Name
		})

		g.writeLn("type %s struct {", t.Name)
		for _, f := range t.Fields {
			g.writeLn("%s %s", f.Name, f.DataType.Decl())
		}
		g.writeLn("}")
		g.writeLn("")
	}

	// Generate a chan type for every stream arg or ret, but only once!
	genChans := make(map[string]struct{})
	for _, srvc := range srvcs {
		for _, s := range srvc.Streams {
			if s.Args != nil {
				if _, ok := genChans[s.Args.Name()]; !ok {
					genChans[s.Args.Name()] = struct{}{}
					g.genChanType(s.Args)
				}
			}
			if s.Ret != nil {
				if _, ok := genChans[s.Ret.Name()]; !ok {
					genChans[s.Ret.Name()] = struct{}{}
					g.genChanType(s.Ret)
				}
			}
		}
	}
}

func (g *generator) genChanType(dt ast.DataType) {
	gen := func(readOnly bool) {
		infix := "Write"
		if readOnly {
			infix = "Read"
		}

		// Type definition.
		g.writeLn("//msgp:ignore %s%sChan", dt.Name(), infix)
		g.writeLn("type %s%sChan struct {", dt.Name(), infix)
		g.writeLn("closer.Closer")
		g.write("C ")
		if readOnly {
			g.write("<-chan ")
		} else {
			g.write("chan<- ")
		}
		g.writeLn("%s", dt.Decl())
		g.writeLn("c chan %s", dt.Decl())
		g.writeLn("mx sync.Mutex")
		g.writeLn("err error")
		g.writeLn("}")
		g.writeLn("")

		// Constructor.
		g.writeLn("func new%s%sChan(cl closer.Closer, size uint) *%s%sChan {", dt.Name(), infix, dt.Name(), infix)
		g.writeLn("c := &%s%sChan{Closer: cl, c: make(chan %s, size)}", dt.Name(), infix, dt.Decl())
		g.writeLn("c.C = c.c")
		g.writeLn("return c")
		g.writeLn("}")
		g.writeLn("")

		// setError method.
		g.writeLn("func (c *%s%sChan) setError(err error) {", dt.Name(), infix)
		g.writeLn("c.mx.Lock()")
		g.writeLn("c.err = err")
		g.writeLn("c.mx.Unlock()")
		g.writeLn("c.Close_()")
		g.writeLn("}")
		g.writeLn("")

		// Err method.
		g.writeLn("func (c *%s%sChan) Err() (err error) {", dt.Name(), infix)
		g.writeLn("c.mx.Lock()")
		g.writeLn("err = c.err")
		g.writeLn("c.mx.Unlock()")
		g.writeLn("return")
		g.writeLn("}")
		g.writeLn("")
	}
	gen(true)
	gen(false)
}
