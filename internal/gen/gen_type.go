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

	"github.com/desertbit/orbit/internal/parse"
)

func (g *generator) genTypes(ts []*parse.Type, srvcs []*parse.Service, streamChanSize uint) {
	g.writeLn("//#############//")
	g.writeLn("//### Types ###//")
	g.writeLn("//#############//")
	g.writeLn("")

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

		g.writeLn("type %s struct {", t.Name)
		for _, f := range t.Fields {
			g.writeLn("%s %s", f.Name, f.DataType.String())
		}
		g.writeLn("}")
		g.writeLn("")

		// Generate a chan type, if it is used in a stream as arg or ret value.
		for _, srvc := range srvcs {
			for _, e := range srvc.Entries {
				if s, ok := e.(*parse.Stream); ok {
					if (s.HasArgs() && s.Args().Name == t.Name) || (s.HasRet() && s.Ret().Name == t.Name) {
						g.genChanType(t.Name, false, streamChanSize)
						g.genChanType(t.Name, true, streamChanSize)
						continue NextType
					}
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
	g.writeLn("//msgp:ignore %s%sChan", name, suffix)
	g.writeLn("type %s%sChan struct {", name, suffix)
	g.writeLn("closer.Closer")
	g.write("C ")
	if ro {
		g.write("<-chan ")
	} else {
		g.write("chan<- ")
	}
	g.writeLn("*%s", name)
	g.writeLn("c chan *%s", name)
	g.writeLn("mx sync.Mutex")
	g.writeLn("err error")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writeLn("func new%s%sChan(cl closer.Closer) *%s%sChan {", name, suffix, name, suffix)
	g.writeLn("c := &%s%sChan{Closer: cl, c: make(chan *%s, %d)}", name, suffix, name, streamChanSize)
	g.writeLn("c.C = c.c")
	g.writeLn("return c")
	g.writeLn("}")
	g.writeLn("")

	// setError method.
	g.writeLn("func (c *%s%sChan) setError(err error) {", name, suffix)
	g.writeLn("c.mx.Lock()")
	g.writeLn("c.err = err")
	g.writeLn("c.mx.Unlock()")
	g.writeLn("c.Close_()")
	g.writeLn("}")
	g.writeLn("")

	// Err method.
	g.writeLn("func (c *%s%sChan) Err() (err error) {", name, suffix)
	g.writeLn("c.mx.Lock()")
	g.writeLn("err = c.err")
	g.writeLn("c.mx.Unlock()")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}
