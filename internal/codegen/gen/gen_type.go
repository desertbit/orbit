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
		g.writefLn("type %s struct {", t.Name)
		for _, f := range t.Fields {
			g.writef("%s %s", f.Name, f.DataType.Decl())
			if f.StructTag != "" {
				g.writef(" `%s`", f.StructTag)
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
				g.genReadStreamType(s.Arg)
				g.genWriteStreamType(s.Arg)
			}
		}
		if s.Ret != nil {
			if _, ok := genStreams[s.Ret.Name()]; !ok {
				genStreams[s.Ret.Name()] = struct{}{}
				g.genReadStreamType(s.Ret)
				g.genWriteStreamType(s.Ret)
			}
		}
	}
}

func (g *generator) genReadStreamType(dt ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sReadStream", dt.Name())
	g.writefLn("type %sReadStream struct {", dt.Name())
	g.writeLn("closer.Closer")
	g.writeLn("stream transport.Stream")
	g.writeLn("codec codec.Codec")
	g.writeLn("maxSize int")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sReadStream(cl closer.Closer, s transport.Stream, cc codec.Codec, ms int) *%sReadStream {", dt.Name(), dt.Name())
	g.writefLn("return &%sReadStream{Closer: cl, stream: s, codec: cc, maxSize: ms}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writefLn("func (%s *%sReadStream) Read() (arg %s, err error) {", recv, dt.Name(), dt.Decl())
	g.writefLn("if %s.IsClosing() {", recv)
	g.writeLn("err = ErrClosed")
	g.writeLn("return")
	g.writeLn("}")
	g.writefLn("arg = %s", dt.ZeroValue())
	g.writefLn("err = packet.ReadDecode(%s.stream, &arg, %s.codec, %s.maxSize)", recv, recv, recv)
	g.writeLn("if err != nil {")
	g.writefLn("if errors.Is(err, packet.ErrZeroData) || errors.Is(err, io.EOF) || %s.stream.IsClosed() {", recv)
	g.writeLn("err = ErrClosed")
	g.writeLn("}")
	g.writefLn("%s.Close_()", recv)
	g.writeLn("return")
	g.writeLn("}")

	// Validate, if needed.
	g.writeValErrCheck(dt, "arg")

	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genWriteStreamType(dt ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sWriteStream", dt.Name())
	g.writefLn("type %sWriteStream struct {", dt.Name())
	g.writeLn("closer.Closer")
	g.writeLn("stream transport.Stream")
	g.writeLn("codec codec.Codec")
	g.writeLn("maxSize int")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sWriteStream(cl closer.Closer, s transport.Stream, cc codec.Codec, ms int) *%sWriteStream {", dt.Name(), dt.Name())
	g.writeLn("cl.OnClosing(func() error { return packet.Write(s, nil, 0) })")
	g.writefLn("return &%sWriteStream{Closer: cl, stream: s, codec: cc, maxSize: ms}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writefLn("func (%s *%sWriteStream) Write(ret %s) (err error) {", recv, dt.Name(), dt.Decl())
	g.writefLn("if %s.IsClosing() {", recv)
	g.writeLn("err = ErrClosed")
	g.writeLn("return")
	g.writeLn("}")
	g.writefLn("err = packet.WriteEncode(%s.stream, ret, %s.codec, %s.maxSize)", recv, recv, recv)
	g.writeLn("if err != nil {")
	g.writefLn("if errors.Is(err, io.EOF) || %s.stream.IsClosed() {", recv)
	g.writefLn("%s.Close_()", recv)
	g.writeLn("return ErrClosed")
	g.writeLn("}")
	g.writeLn("}")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}
