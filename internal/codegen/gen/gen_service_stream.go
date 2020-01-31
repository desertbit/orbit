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

func (g *generator) genServiceStreamCallerSignature(s *ast.Stream) {
	g.write("%s(ctx context.Context) (", s.Name)
	if s.Args != nil {
		g.write("args *%sWriteChan, ", s.Args.Name())
	}
	if s.Ret != nil {
		g.write("ret *%sReadChan, ", s.Ret.Name())
	} else if s.Args == nil {
		g.write("stream net.Conn, ")
	}
	g.write("err error)")
}

func (g *generator) genServiceStreamCaller(s *ast.Stream, srvcName, structName string, errs []*ast.Error) {
	// Method declaration.
	g.write("func (%s *%s) ", recv, structName)
	g.genServiceStreamCallerSignature(s)
	g.writeLn(" {")

	// Method body.
	// First, open the stream.
	if s.Args == nil && s.Ret == nil {
		g.writeLn("return %s.s.OpenStream(ctx, Service%s, %s)", recv, srvcName, srvcName+s.Name)
		g.writeLn("}")
		g.writeLn("")
		return
	}

	g.writeLn("stream, err := %s.s.OpenStream(ctx, Service%s, %s)", recv, srvcName, srvcName+s.Name)
	g.errIfNil()

	if s.Args != nil {
		g.writeLn("args = new%sWriteChan(%s.s.CloserOneWay(), stream, %s.s.Codec())", s.Args.Name(), recv, recv)

		// Close unidirectional stream immediately.
		if s.Ret == nil {
			g.writeLn("args.OnClosing(stream.Close)")
		}
	}

	if s.Ret != nil {
		g.writeLn("ret = new%sReadChan(%s.s.CloserOneWay(), stream, %s.s.Codec())", s.Ret.Name(), recv, recv)

		// Close unidirectional stream immediately.
		if s.Args == nil {
			g.writeLn("ret.OnClosing(stream.Close)")
		}
	}

	// For a bidirectional stream, we need a new goroutine.
	if s.Args != nil && s.Ret != nil {
		g.writeLn("go func() {")
		g.writeLn("<-args.ClosedChan()")
		g.writeLn("<-ret.ClosedChan()")
		g.writeLn("_ = stream.Close()")
		g.writeLn("}()")
	}

	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceStreamHandlerSignature(s *ast.Stream) {
	g.write("%s(s *orbit.Session, ", s.Name)
	if s.Args != nil {
		g.write("args *%sReadChan, ", s.Args.Name())
	}
	if s.Ret != nil {
		g.write("ret *%sWriteChan", s.Ret.Name())
	} else if s.Args == nil {
		g.write("stream net.Conn")
	}
	g.write(")")
}

func (g *generator) genServiceStreamHandler(s *ast.Stream, structName string, errs []*ast.Error) {
	// Method declaration.
	g.writeLn("func (%s *%s) %s(s *orbit.Session, stream net.Conn) {", recv, structName, s.NamePrv())

	if s.Args == nil && s.Ret == nil {
		g.writeLn("%s.h.%s(s, stream)", recv, s.Name)
		g.writeLn("}")
		return
	}

	handlerArgs := "s, "
	if s.Args != nil {
		handlerArgs += "args, "
		g.writeLn("args := new%sReadChan(%s.s.CloserOneWay(), stream, %s.s.Codec())", s.Args.Name(), recv, recv)

		// Close unidirectional stream immediately.
		if s.Ret == nil {
			g.writeLn("args.OnClosing(stream.Close)")
		}
	}
	if s.Ret != nil {
		handlerArgs += "ret"
		g.writeLn("ret := new%sWriteChan(%s.s.CloserOneWay(), stream, %s.s.Codec())", s.Ret.Name(), recv, recv)

		// Close unidirectional stream immediately.
		if s.Args == nil {
			g.writeLn("ret.OnClosing(stream.Close)")
		}
	}

	// For a bidirectional stream, we need a new goroutine.
	if s.Args != nil && s.Ret != nil {
		g.writeLn("go func() {")
		g.writeLn("<-args.ClosedChan()")
		g.writeLn("<-ret.ClosedChan()")
		g.writeLn("_ = stream.Close()")
		g.writeLn("}()")
	}

	g.writeLn("%s.h.%s(%s)", recv, s.Name, handlerArgs)
	g.writeLn("}")
	g.writeLn("")
}
