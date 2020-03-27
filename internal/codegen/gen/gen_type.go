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

func (g *generator) genTypes(ts []*ast.Type, srvc *ast.Service) {
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

	// Generate a stream type for every stream arg or ret, but only once!
	genStreams := make(map[string]struct{})
	for _, s := range srvc.Streams {
		if s.Arg != nil {
			if _, ok := genStreams[s.Arg.Name()]; !ok {
				genStreams[s.Arg.Name()] = struct{}{}
				// Check, if the data type is a validation
				g.genReadStreamType(s.Arg, s.ArgValTag)
				g.genWriteStreamType(s.Arg)
			}
		}
		if s.Ret != nil {
			if _, ok := genStreams[s.Ret.Name()]; !ok {
				genStreams[s.Ret.Name()] = struct{}{}
				g.genReadStreamType(s.Ret, s.RetValTag)
				g.genWriteStreamType(s.Ret)
			}
		}
	}
}

func (g *generator) genReadStreamType(dt ast.DataType, valTag string) {
	// Type definition.
	g.writeLn("//msgp:ignore %sReadStream", dt.Name())
	g.writeLn("type %sReadStream struct {", dt.Name())
	g.writeLn("closer.Closer")
	g.writeLn("stream net.Conn")
	g.writeLn("codec codec.Codec")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writeLn("func new%sReadStream(cl closer.Closer, stream net.Conn, cc codec.Codec) *%sReadStream {", dt.Name(), dt.Name())
	g.writeLn("return &%sReadStream{Closer: cl, stream: stream, codec: cc}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writeLn("func (c *%sReadStream) Read() (arg %s, err error) {", dt.Name(), dt.Decl())
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

func (g *generator) genWriteStreamType(dt ast.DataType) {
	// Type definition.
	g.writeLn("//msgp:ignore %sWriteStream", dt.Name())
	g.writeLn("type %sWriteStream struct {", dt.Name())
	g.writeLn("closer.Closer")
	g.writeLn("stream net.Conn")
	g.writeLn("codec codec.Codec")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writeLn("func new%sWriteStream(cl closer.Closer, stream net.Conn, cc codec.Codec) *%sWriteStream {", dt.Name(), dt.Name())
	g.writeLn("cl.OnClosing(func() error { return packet.Write(stream, nil) })")
	g.writeLn("return &%sWriteStream{Closer: cl, stream: stream, codec: cc}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writeLn("func (c *%sWriteStream) Write(ret %s) (err error) {", dt.Name(), dt.Decl())
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
