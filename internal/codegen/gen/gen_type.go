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

const (
	streamErrCode = "_streamErrCode"
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

	// Generate a stream type for every stream arg or ret or bidirectional stream func,
	// but only once!
	genStreams := make(map[string]struct{})
	for _, s := range srvc.Streams {
		if s.Arg == nil && s.Ret == nil {
			continue
		}

		if s.Arg != nil && s.Ret != nil {
			// Bidirectional stream.
			if _, ok := genStreams[s.Name]; !ok {
				genStreams[s.Name] = struct{}{}
				g.genClientStreamType(s.Name, s.Arg, s.Ret)
				g.genServiceStreamType(s.Name, s.Arg, s.Ret)
			}
		} else {
			if s.Arg != nil {
				if _, ok := genStreams[s.Arg.Name()]; !ok {
					genStreams[s.Arg.Name()] = struct{}{}
					g.genServiceReadStreamType(s.Arg)
					g.genClientWriteStreamType(s.Arg)
				}
			}
			if s.Ret != nil {
				if _, ok := genStreams[s.Ret.Name()]; !ok {
					genStreams[s.Ret.Name()] = struct{}{}
					g.genClientReadStreamType(s.Ret)
					g.genServiceWriteStreamType(s.Ret)
				}
			}
		}
	}
}

// genClientReadStreamType generates the client read side of a stream.
func (g *generator) genClientReadStreamType(dt ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sReadStream", dt.Name())
	g.writefLn("type %sReadStream struct {", dt.Name())
	g.writeLn("stream oclient.TypedRStream")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sReadStream(s oclient.TypedRStream) *%sReadStream {", dt.Name(), dt.Name())
	g.writefLn("return &%sReadStream{stream: s}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writefLn("func (%s *%sReadStream) Read() (ret %s, err error) {", recv, dt.Name(), dt.Decl())
	g.writefLn("err = %s.stream.Read(&ret)", recv)
	g.errIfNilFunc(func() {
		g.writefLn("err = %s(err)", clientErrorCheck)
		g.writeLn("if errors.Is(err, oclient.ErrClosed) {")
		g.writeLn("err = ErrClosed")
		g.writeLn("}")
		g.writeLn("return")
	})
	// Validate, if needed.
	g.writeValErrCheck(dt, "ret")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

// genClientWriteStreamType generates the client write side of a stream.
func (g *generator) genClientWriteStreamType(dt ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sWriteStream", dt.Name())
	g.writefLn("type %sWriteStream struct {", dt.Name())
	g.writeLn("stream oclient.TypedWStream")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sWriteStream(s oclient.TypedWStream) *%sWriteStream {", dt.Name(), dt.Name())
	g.writefLn("return &%sWriteStream{stream: s}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writefLn("func (%s *%sWriteStream) Write(arg %s) (err error) {", recv, dt.Name(), dt.Decl())
	g.writefLn("err = %s.stream.Write(arg)", recv)
	g.errIfNilFunc(func() {
		g.writefLn("err = %s(err)", clientErrorCheck)
		g.writeLn("if errors.Is(err, oclient.ErrClosed) {")
		g.writeLn("err = ErrClosed")
		g.writeLn("}")
		g.writeLn("return")
	})
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

// genClientStreamType generates the client side of a bidirectional stream.
func (g *generator) genClientStreamType(name string, arg, ret ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sClientStream", name)
	g.writefLn("type %sClientStream struct {", name)
	g.writeLn("stream oclient.TypedRWStream")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sClientStream(s oclient.TypedRWStream) *%sClientStream {", name, name)
	g.writefLn("return &%sClientStream{stream: s}", name)
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writefLn("func (%s *%sClientStream) Read() (ret %s, err error) {", recv, name, ret.Decl())
	g.writefLn("err = %s.stream.Read(&ret)", recv)
	g.errIfNilFunc(func() {
		g.writefLn("err = %s(err)", clientErrorCheck)
		g.writeLn("if errors.Is(err, oclient.ErrClosed) {")
		g.writeLn("err = ErrClosed")
		g.writeLn("}")
		g.writeLn("return")
	})
	// Validate, if needed.
	g.writeValErrCheck(ret, "ret")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writefLn("func (%s *%sClientStream) Write(arg %s) (err error) {", recv, name, arg.Decl())
	g.writefLn("err = %s.stream.Write(arg)", recv)
	g.errIfNilFunc(func() {
		g.writefLn("err = %s(err)", clientErrorCheck)
		g.writeLn("if errors.Is(err, oclient.ErrClosed) {")
		g.writeLn("err = ErrClosed")
		g.writeLn("}")
		g.writeLn("return")
	})
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

// genServiceReadStreamType generates the service read side of a stream.
func (g *generator) genServiceReadStreamType(dt ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sReadStream", dt.Name())
	g.writefLn("type %sReadStream struct {", dt.Name())
	g.writeLn("stream oservice.TypedRStream")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sReadStream(s oservice.TypedRStream) *%sReadStream {", dt.Name(), dt.Name())
	g.writefLn("return &%sReadStream{stream: s}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writefLn("func (%s *%sReadStream) Read() (arg %s, err error) {", recv, dt.Name(), dt.Decl())
	g.writefLn("err = %s.stream.Read(&arg)", recv)
	g.errIfNilFunc(func() {
		g.writefLn("err = %s(err)", serviceErrorCheck)
		g.writeLn("if errors.Is(err, oservice.ErrClosed) {")
		g.writeLn("err = ErrClosed")
		g.writeLn("}")
		g.writeLn("return")
	})
	// Validate, if needed.
	g.writeValErrCheck(dt, "arg")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

// genServiceWriteStreamType generates the service write side of a stream.
func (g *generator) genServiceWriteStreamType(dt ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sWriteStream", dt.Name())
	g.writefLn("type %sWriteStream struct {", dt.Name())
	g.writeLn("stream oservice.TypedWStream")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sWriteStream(s oservice.TypedWStream) *%sWriteStream {", dt.Name(), dt.Name())
	g.writefLn("return &%sWriteStream{stream: s}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writefLn("func (%s *%sWriteStream) Write(ret %s) (err error) {", recv, dt.Name(), dt.Decl())
	g.writefLn("err = %s.stream.Write(ret)", recv)
	g.errIfNilFunc(func() {
		g.writefLn("err = %s(err)", serviceErrorCheck)
		g.writeLn("if errors.Is(err, oservice.ErrClosed) {")
		g.writeLn("err = ErrClosed")
		g.writeLn("}")
		g.writeLn("return")
	})
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

// genServiceStreamType generates the service side of a bidirectional stream.
func (g *generator) genServiceStreamType(name string, arg, ret ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sServiceStream", name)
	g.writefLn("type %sServiceStream struct {", name)
	g.writeLn("stream oservice.TypedRWStream")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sServiceStream(s oservice.TypedRWStream) *%sServiceStream {", name, name)
	g.writefLn("return &%sServiceStream{stream: s}", name)
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writefLn("func (%s *%sServiceStream) Read() (arg %s, err error) {", recv, name, arg.Decl())
	g.writefLn("err = %s.stream.Read(&arg)", recv)
	g.errIfNilFunc(func() {
		g.writefLn("err = %s(err)", serviceErrorCheck)
		g.writeLn("if errors.Is(err, oservice.ErrClosed) {")
		g.writeLn("err = ErrClosed")
		g.writeLn("}")
		g.writeLn("return")
	})
	// Validate, if needed.
	g.writeValErrCheck(arg, "arg")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writefLn("func (%s *%sServiceStream) Write(ret %s) (err error) {", recv, name, ret.Decl())
	g.writefLn("err = %s.stream.Write(ret)", recv)
	g.errIfNilFunc(func() {
		g.writefLn("err = %s(err)", serviceErrorCheck)
		g.writeLn("if errors.Is(err, oservice.ErrClosed) {")
		g.writeLn("err = ErrClosed")
		g.writeLn("}")
		g.writeLn("return")
	})
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}
