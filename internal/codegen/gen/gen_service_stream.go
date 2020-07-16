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
	"github.com/desertbit/orbit/internal/codegen/ast"
)

//##############//
//### Client ###//
//##############//

func (g *generator) genClientStreamSignature(s *ast.Stream) {
	if s.Arg == nil && s.Ret == nil {
		// Raw.
		g.writef("%s(ctx context.Context) (stream transport.Stream, err error)", s.Name)
	} else {
		// Typed.
		g.writef("%s(ctx context.Context) (stream *%sClientStream, err error)", s.Name, s.Name)
	}
}

func (g *generator) genClientStream(s *ast.Stream, errs []*ast.Error) {
	// Method declaration.
	g.writef("func (%s *client) ", recv)
	g.genClientStreamSignature(s)
	g.writeLn(" {")

	// Ensure Timeout on context.
	g.writefLn("if %s.streamInitTimeout > 0 {", recv)
	g.writeLn("var cancel context.CancelFunc")
	g.writefLn("ctx, cancel = context.WithTimeout(ctx, %s.streamInitTimeout)", recv)
	g.writeLn("defer cancel()")
	g.writeLn("}")

	// Implementation.
	if s.Arg == nil && s.Ret == nil {
		// Raw.
		g.writefLn("stream, err = %s.Stream(ctx, StreamID%s)", recv, s.Name)
		g.errIfNil()
	} else {
		// Typed.
		g.writef("str, err := %s.%s(ctx, StreamID%s,", recv, typedStream(s, false), s.Name)
		if s.Arg != nil {
			g.writeOrbitMaxSizeParam(s.MaxArgSize, false)
		}
		if s.Ret != nil {
			g.writeOrbitMaxSizeParam(s.MaxRetSize, false)
		}
		g.writeLn(")")
		g.errIfNil()
		g.writefLn("stream = new%sClientStream(str)", s.Name)
	}

	// End of method.
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

//###############//
//### Service ###//
//###############//

func (g *generator) genServiceStreamRegister(s *ast.Stream) {
	if s.Arg == nil && s.Ret == nil {
		// Raw.
		g.writefLn("os.RegisterStream(StreamID%s, srvc.%s)", s.Name, s.NamePrv())
	} else {
		// Typed.
		g.writef("os.Register%s(StreamID%s, srvc.%s,", typedStream(s, true), s.Name, s.NamePrv())
		if s.Arg != nil {
			g.writeOrbitMaxSizeParam(s.MaxArgSize, true)
		}
		if s.Ret != nil {
			g.writeOrbitMaxSizeParam(s.MaxRetSize, true)
		}
		g.writeLn(")")
	}
}

func (g *generator) genServiceHandlerStreamSignature(s *ast.Stream) {
	g.writef("%s(ctx oservice.Context, ", s.Name)

	if s.Arg == nil && s.Ret == nil {
		// Raw.
		g.writeLn("stream transport.Stream)")
	} else {
		// Typed.
		g.writefLn("stream *%sServiceStream) error", s.Name)
	}
}

func (g *generator) genServiceStream(s *ast.Stream) {
	g.writef("func (%s *service) %s(ctx oservice.Context, ", recv, s.NamePrv())

	if s.Arg == nil && s.Ret == nil {
		// Raw.
		g.writeLn("stream transport.Stream) {")
		g.writefLn("%s.h.%s(ctx, stream)", recv, s.Name)
	} else {
		// Typed.
		g.writefLn("stream oservice.%s) error {", typedStream(s, true))
		g.writefLn("return %s.h.%s(ctx, new%sServiceStream(stream))", recv, s.Name, s.Name)
	}

	g.writeLn("}")
	g.writeLn("")
}

//###############//
//### Private ###//
//###############//

func typedStream(s *ast.Stream, service bool) (dataType string) {
	if s.Arg != nil && s.Ret != nil {
		// ReadWrite.
		return "TypedRWStream"
	} else if (!service && s.Ret != nil) || (service && s.Arg != nil) {
		// Read.
		return "TypedRStream"
	} else {
		// Write.
		return "TypedWStream"
	}
}
