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
		g.writefLn("type %s struct {", t.Ident())
		for _, f := range t.Fields {
			g.writef("%s %s", f.Ident(), f.DataType.Decl())
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
	for _, s := range srvc.Streams {
		g.genClientStreamType(s)
		g.genServiceStreamType(s)
	}
}

// genClientStreamType generates the client read side of a stream.
func (g *generator) genClientStreamType(s *ast.Stream) {
	if s.Arg == nil && s.Ret == nil {
		// Raw streams do not need a stream type.
		return
	}

	name := s.Ident() + "ClientStream"
	typedStream := "oclient." + typedStream(s, false)

	// Type definition.
	g.writefLn("//msgp:ignore %s", name)
	g.writefLn("type %s struct {", name)
	g.writefLn("oclient.TypedStreamCloser")
	g.writefLn("stream %s", typedStream)
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%s(s %s) *%s {", name, typedStream, name)
	g.writefLn("return &%s{TypedStreamCloser: s, stream: s}", name)
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	if s.Ret != nil {
		g.writefLn("func (%s *%s) Read() (ret %s, err error) {", recv, name, s.Ret.Decl())
		g.writefLn("err = %s.stream.Read(&ret)", recv)
		g.errIfNilFunc(func() {
			g.writeLn("if errors.Is(err, oclient.ErrClosed) {")
			g.writeLn("err = ErrClosed")
			g.writeLn("return")
			g.writeLn("}")
			if len(s.Errors) != 0 {
				// Inline check for defined errors.
				g.genClientErrorInlineCheck(s.Errors)
			} else {
				g.writeLn("return")
			}
		})
		// Validate, if needed.
		g.writeValErrCheck(s.Ret, "ret")
		g.writeLn("return")
		g.writeLn("}")
		g.writeLn("")
	}

	// Write method.
	if s.Arg != nil {
		g.writefLn("func (%s *%s) Write(arg %s) (err error) {", recv, name, s.Arg.Decl())
		g.writefLn("err = %s.stream.Write(arg)", recv)
		g.errIfNilFunc(func() {
			g.writeLn("if errors.Is(err, oclient.ErrClosed) {")
			g.writeLn("err = ErrClosed")
			g.writeLn("return")
			g.writeLn("}")
			if len(s.Errors) != 0 {
				// Inline check for defined errors.
				g.genClientErrorInlineCheck(s.Errors)
			} else {
				g.writeLn("return")
			}
		})
		g.writeLn("return")
		g.writeLn("}")
		g.writeLn("")
	}
}

func (g *generator) genServiceStreamType(s *ast.Stream) {
	if s.Arg == nil && s.Ret == nil {
		// Raw streams do not need a stream type.
		return
	}

	name := s.Ident() + "ServiceStream"
	typedStream := "oservice." + typedStream(s, true)

	// Type definition.
	g.writefLn("//msgp:ignore %s", name)
	g.writefLn("type %s struct {", name)
	g.writefLn("oservice.TypedStreamCloser")
	g.writefLn("stream %s", typedStream)
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%s(s %s) *%s {", name, typedStream, name)
	g.writefLn("return &%s{TypedStreamCloser: s, stream: s}", name)
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	if s.Arg != nil {
		g.writefLn("func (%s *%s) Read() (arg %s, err error) {", recv, name, s.Arg.Decl())
		g.writefLn("err = %s.stream.Read(&arg)", recv)
		g.errIfNilFunc(func() {
			g.writeLn("if errors.Is(err, oservice.ErrClosed) {")
			g.writeLn("err = ErrClosed")
			g.writeLn("return")
			g.writeLn("}")
			if len(s.Errors) != 0 {
				// Inline check for defined errors.
				g.genServiceErrorInlineCheck(s.Errors)
			} else {
				g.writeLn("return")
			}
		})
		// Validate, if needed.
		g.writeValErrCheck(s.Arg, "arg")
		g.writeLn("return")
		g.writeLn("}")
		g.writeLn("")
	}

	// Write method.
	if s.Ret != nil {
		g.writefLn("func (%s *%s) Write(ret %s) (err error) {", recv, name, s.Ret.Decl())
		g.writefLn("err = %s.stream.Write(ret)", recv)
		g.errIfNilFunc(func() {
			g.writeLn("if errors.Is(err, oservice.ErrClosed) {")
			g.writeLn("err = ErrClosed")
			g.writeLn("return")
			g.writeLn("}")
			if len(s.Errors) != 0 {
				// Inline check for defined errors.
				g.genServiceErrorInlineCheck(s.Errors)
			} else {
				g.writeLn("return")
			}
		})
		g.writeLn("return")
		g.writeLn("}")
		g.writeLn("")
	}
}
