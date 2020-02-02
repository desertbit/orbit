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
		g.writeLn("type %s struct {", t.Name)
		for _, f := range t.Fields {
			g.write("%s %s", f.Name, f.DataType.Decl())
			if f.ValTag != "" {
				g.write(" `validate:\"%s\"`", f.ValTag)
			}
			g.writeLn("")
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
					// Check, if the data type is a validation
					g.genReadChanType(s.Args, s.ArgsValTag)
					g.genWriteChanType(s.Args)
				}
			}
			if s.Ret != nil {
				if _, ok := genChans[s.Ret.Name()]; !ok {
					genChans[s.Ret.Name()] = struct{}{}
					g.genReadChanType(s.Ret, s.RetValTag)
					g.genWriteChanType(s.Ret)
				}
			}
		}
	}
}

func (g *generator) genReadChanType(dt ast.DataType, valTag string) {
	// Type definition.
	g.writeLn("//msgp:ignore %sReadChan", dt.Name())
	g.writeLn("type %sReadChan struct {", dt.Name())
	g.writeLn("closer.Closer")
	g.writeLn("stream net.Conn")
	g.writeLn("codec codec.Codec")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writeLn("func new%sReadChan(cl closer.Closer, stream net.Conn, cc codec.Codec) *%sReadChan {", dt.Name(), dt.Name())
	g.writeLn("return &%sReadChan{Closer: cl, stream: stream, codec: cc}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writeLn("func (c *%sReadChan) Read() (arg %s, err error) {", dt.Name(), dt.Decl())
	g.writeLn("if c.IsClosing() {")
	g.writeLn("err = ErrClosed")
	g.writeLn("return")
	g.writeLn("}")
	// Parse.
	g.writeLn("err = packet.ReadDecode(c.stream, &arg, c.codec)")
	g.writeLn("if err != nil {")
	g.writeLn("if errors.Is(err, packet.ErrZeroData) || errors.Is(err, io.EOF) {")
	g.writeLn("err = ErrClosed")
	g.writeLn("}")
	g.writeLn("c.Close_()")
	g.writeLn("return")
	g.writeLn("}")
	// Validate.
	if valTag != "" {
		// Validate a single value.
		g.writeLn("err = validate.Var(arg, \"%s\")", valTag)
	} else {
		// Validate a struct.
		g.writeLn("err = validate.Struct(arg)")
	}
	g.errIfNilFunc(func() {
		g.writeLn("err = %s(err)", valErrorCheck)
		g.writeLn("return")
	})
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genWriteChanType(dt ast.DataType) {
	// Type definition.
	g.writeLn("//msgp:ignore %sWriteChan", dt.Name())
	g.writeLn("type %sWriteChan struct {", dt.Name())
	g.writeLn("closer.Closer")
	g.writeLn("stream net.Conn")
	g.writeLn("codec codec.Codec")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writeLn("func new%sWriteChan(cl closer.Closer, stream net.Conn, cc codec.Codec) *%sWriteChan {", dt.Name(), dt.Name())
	g.writeLn("cl.OnClosing(func() error { return packet.Write(stream, nil) })")
	g.writeLn("return &%sWriteChan{Closer: cl, stream: stream, codec: cc}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writeLn("func (c *%sWriteChan) Write(ret %s) (err error) {", dt.Name(), dt.Decl())
	g.writeLn("if c.IsClosing() {")
	g.writeLn("err = ErrClosed")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("err = packet.WriteEncode(c.stream, ret, c.codec)")
	g.writeLn("if err != nil {")
	g.writeLn("if errors.Is(err, io.EOF) {")
	g.writeLn("c.Close_()")
	g.writeLn("return ErrClosed")
	g.writeLn("}")
	g.writeLn("}")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}
