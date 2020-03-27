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

func (g *generator) genServiceClientStreamSignature(s *ast.Stream) {
	g.write("%s(ctx context.Context) (", s.Name)
	if s.Arg != nil {
		g.write("arg *%sWriteStream, ", s.Arg.Name())
	}
	if s.Ret != nil {
		g.write("ret *%sReadStream, ", s.Ret.Name())
	} else if s.Arg == nil {
		g.write("stream transport.Stream, ")
	}
	g.write("err error)")
}

func (g *generator) genServiceClientStream(s *ast.Stream, errs []*ast.Error) {
	// Method declaration.
	g.write("func (%s *client) ", recv)
	g.genServiceClientStreamSignature(s)
	g.writeLn(" {")

	// Method body.
	g.writeLn("if %s.streamInitTimeout > 0 {", recv)
	g.writeLn("var cancel context.CancelFunc")
	g.writeLn("ctx, cancel = context.WithTimeout(ctx, %s.streamInitTimeout)", recv)
	g.writeLn("defer cancel()")
	g.writeLn("}")

	g.write("stream, err ")
	if s.Arg == nil && s.Ret == nil {
		g.write(" = ")
	} else {
		g.write(" := ")
	}
	g.writeLn("%s.Stream(ctx, %s)", recv, s.Name)
	g.errIfNil()

	if s.Arg != nil {
		g.writeLn("arg = new%sWriteStream(%s.CloserOneWay(), stream, %s.codec)", s.Arg.Name(), recv, recv)

		// Close unidirectional stream immediately.
		if s.Ret == nil {
			g.writeLn("arg.OnClosing(stream.Close)")
		}
	}

	if s.Ret != nil {
		g.writeLn("ret = new%sReadStream(%s.CloserOneWay(), stream, %s.codec)", s.Ret.Name(), recv, recv)

		// Close unidirectional stream immediately.
		if s.Arg == nil {
			g.writeLn("ret.OnClosing(stream.Close)")
		}
	}

	// For a bidirectional stream, we need a new goroutine.
	if s.Arg != nil && s.Ret != nil {
		g.writeLn("go func() {")
		g.writeLn("<-arg.ClosedChan()")
		g.writeLn("<-ret.ClosedChan()")
		g.writeLn("_ = stream.Close()")
		g.writeLn("}()")
	}

	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceHandlerStreamSignature(s *ast.Stream) {
	g.write("%s(ctx oservice.Context,", s.Name)
	if s.Arg != nil {
		g.write("arg *%sReadStream,", s.Arg.Name())
	}
	if s.Ret != nil {
		g.write("ret *%sWriteStream,", s.Ret.Name())
	} else if s.Arg == nil {
		g.write("stream transport.Stream,")
	}
	g.write(")")
}

func (g *generator) genServiceHandlerStream(s *ast.Stream, errs []*ast.Error) {
	// Method declaration.
	g.writeLn("func (%s *service) %s(ctx oservice.Context, stream transport.Stream) {", recv, s.NamePrv())

	if s.Arg == nil && s.Ret == nil {
		g.writeLn("%s.h.%s(ctx, stream)", recv, s.Name)
		g.writeLn("}")
		return
	}

	handlerArgs := "ctx,"
	if s.Arg != nil {
		handlerArgs += "arg,"
		g.writeLn("arg := new%sReadStream(%s.CloserOneWay(), stream, %s.codec)", s.Arg.Name(), recv, recv)

		// Close unidirectional stream immediately.
		if s.Ret == nil {
			g.writeLn("arg.OnClosing(stream.Close)")
		}
	}
	if s.Ret != nil {
		handlerArgs += "ret,"
		g.writeLn("ret := new%sWriteStream(%s.CloserOneWay(), stream, %s.codec)", s.Ret.Name(), recv, recv)

		// Close unidirectional stream immediately.
		if s.Arg == nil {
			g.writeLn("ret.OnClosing(stream.Close)")
		}
	}

	// For a bidirectional stream, we need a new goroutine.
	if s.Arg != nil && s.Ret != nil {
		g.writeLn("go func() {")
		g.writeLn("<-arg.ClosedChan()")
		g.writeLn("<-ret.ClosedChan()")
		g.writeLn("_ = stream.Close()")
		g.writeLn("}()")
	}

	g.writeLn("%s.h.%s(%s)", recv, s.Name, handlerArgs)
	g.writeLn("}")
	g.writeLn("")
}
