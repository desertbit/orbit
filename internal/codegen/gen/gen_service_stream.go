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
	"fmt"
	"strconv"

	"github.com/desertbit/orbit/internal/codegen/ast"
)

func (g *generator) genServiceClientStreamSignature(s *ast.Stream) {
	g.writef("%s(ctx context.Context) (", s.Name)
	if s.Arg != nil {
		g.writef("arg *%sWriteStream, ", s.Arg.Name())
	}
	if s.Ret != nil {
		g.writef("ret *%sReadStream, ", s.Ret.Name())
	} else if s.Arg == nil {
		g.write("stream transport.Stream, ")
	}
	g.write("err error)")
}

func (g *generator) genServiceClientStream(s *ast.Stream, errs []*ast.Error) {
	// Method declaration.
	g.writef("func (%s *client) ", recv)
	g.genServiceClientStreamSignature(s)
	g.writeLn(" {")

	// Method body.
	g.writefLn("if %s.streamInitTimeout > 0 {", recv)
	g.writeLn("var cancel context.CancelFunc")
	g.writefLn("ctx, cancel = context.WithTimeout(ctx, %s.streamInitTimeout)", recv)
	g.writeLn("defer cancel()")
	g.writeLn("}")

	g.write("stream, err ")
	if s.Arg == nil && s.Ret == nil {
		g.write(" = ")
	} else {
		g.write(" := ")
	}
	g.writefLn("%s.Stream(ctx, %s)", recv, s.Name)
	g.errIfNil()

	if s.Arg != nil {
		g.writef("arg = new%sWriteStream(%s.CloserOneWay(), stream, %s.codec,", s.Arg.Name(), recv, recv)
		g.writePacketMaxSizeParam(s.MaxArgSize, fmt.Sprintf("%s.maxArgSize", recv))
		g.writeLn(")")

		// Close unidirectional stream immediately.
		if s.Ret == nil {
			g.writeLn("arg.OnClosing(stream.Close)")
		}
	}

	if s.Ret != nil {
		g.writef("ret = new%sReadStream(%s.CloserOneWay(), stream, %s.codec,", s.Ret.Name(), recv, recv)
		g.writePacketMaxSizeParam(s.MaxRetSize, fmt.Sprintf("%s.maxRetSize", recv))
		g.writeLn(")")

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
	g.writef("%s(ctx oservice.Context,", s.Name)
	if s.Arg != nil {
		g.writef("arg *%sReadStream,", s.Arg.Name())
	}
	if s.Ret != nil {
		g.writef("ret *%sWriteStream,", s.Ret.Name())
	} else if s.Arg == nil {
		g.write("stream transport.Stream,")
	}
	g.write(")")
}

func (g *generator) genServiceHandlerStream(s *ast.Stream, errs []*ast.Error) {
	// Method declaration.
	g.writefLn("func (%s *service) %s(ctx oservice.Context, stream transport.Stream) {", recv, s.NamePrv())

	if s.Arg == nil && s.Ret == nil {
		g.writefLn("%s.h.%s(ctx, stream)", recv, s.Name)
		g.writeLn("}")
		return
	}

	handlerArgs := "ctx,"
	if s.Arg != nil {
		handlerArgs += "arg,"
		g.writef("arg := new%sReadStream(%s.CloserOneWay(), stream, %s.codec,", s.Arg.Name(), recv, recv)
		g.writePacketMaxSizeParam(s.MaxArgSize, fmt.Sprintf("%s.maxArgSize", recv))
		g.writeLn(")")

		// Close unidirectional stream immediately.
		if s.Ret == nil {
			g.writeLn("arg.OnClosing(stream.Close)")
		}
	}
	if s.Ret != nil {
		handlerArgs += "ret,"
		g.writef("ret := new%sWriteStream(%s.CloserOneWay(), stream, %s.codec,", s.Ret.Name(), recv, recv)
		g.writePacketMaxSizeParam(s.MaxRetSize, fmt.Sprintf("%s.maxRetSize", recv))
		g.writeLn(")")

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

	g.writefLn("%s.h.%s(%s)", recv, s.Name, handlerArgs)
	g.writeLn("}")
	g.writeLn("")
}

// writePacketMaxSizeParam is a helper to determine which max size param must be written
// based on the given params. It automatically handles the special cases
// like no max size or default max size.
// This method must only be used where Packet max size syntax is required.
func (g *generator) writePacketMaxSizeParam(maxSize *int64, defSize string) {
	if maxSize != nil {
		if *maxSize == -1 {
			g.write("packet.NoPayloadSizeLimit")
		} else {
			g.write(strconv.FormatInt(*maxSize, 10))
		}
	} else {
		g.write(defSize)
	}
	g.write(",")
}
