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
	"github.com/desertbit/orbit/internal/parse"
)

func (g *generator) genServiceStreamClient(s *parse.Stream, structName, srvcName string, errs []*parse.Error) {
	// Method declaration.
	g.write("func (%s *%s) ", recv, structName)
	g.write("%s(ctx context.Context) (", s.NamePub())
	if s.HasArgs() {
		g.write("args %sWriteChan, ", s.Args().String())
	}
	if s.HasRet() {
		g.write("ret %sReadChan, ", s.Ret().String())
	} else if !s.HasArgs() {
		g.write("stream net.Conn, ")
	}
	g.write("err error)")
	g.writeLn(" {")

	// Method body.
	// First, open the stream.
	if !s.HasArgs() && !s.HasRet() {
		g.writeLn("return %s.s.OpenStream(ctx, %s, %s)", recv, srvcName, srvcName+s.NamePub())
		g.writeLn("}")
		g.writeLn("")
		return
	}

	g.writeLn("stream, err := %s.s.OpenStream(ctx, %s, %s)", recv, srvcName, srvcName+s.NamePub())
	g.errIfNil()

	if s.HasArgs() {
		g.writeLn("args = new%sWriteChan(%s.s.CloserOneWay())", s.Args().Name, recv)
		g.writeLn("args.OnClosing(func() error { return stream.Close() })")
		g.writeLn("go func() {")
		g.writeLn("closingChan := args.ClosingChan()")
		g.writeLn("codec := %s.s.Codec()", recv)
		g.writeLn("for {")
		g.writeLn("select {")
		g.writeLn("case <- closingChan:")
		g.writeLn("return")
		g.writeLn("case arg := <-args.c:")
		g.writeLn("err := packet.WriteEncode(stream, arg, codec)")
		g.writeLn("if err != nil {")
		g.writeLn("if args.IsClosing() { err = nil }")
		g.writeLn("args.setError(err)")
		g.writeLn("return")
		g.writeLn("}")
		g.writeLn("}")
		g.writeLn("}")
		g.writeLn("}()")
	}

	if s.HasRet() {
		g.writeLn("ret = new%sReadChan(%s.s.CloserOneWay())", s.Ret().Name, recv)
		g.writeLn("ret.OnClosing(func() error { return stream.Close() })")
		g.writeLn("go func() {")
		g.writeLn("closingChan := ret.ClosingChan()")
		g.writeLn("codec := %s.s.Codec()", recv)
		g.writeLn("for {")
		g.writeLn("data := &%s{}", s.Ret().Name)
		g.writeLn("err := packet.ReadDecode(stream, data, codec)")
		g.writeLn("if err != nil {")
		g.writeLn("if ret.IsClosing() { err = nil }")
		g.writeLn("ret.setError(err)")
		g.writeLn("return")
		g.writeLn("}")
		g.writeLn("select {")
		g.writeLn("case <-closingChan:")
		g.writeLn("return")
		g.writeLn("case ret.c <- data:")
		g.writeLn("}")
		g.writeLn("}")
		g.writeLn("}()")
	}

	// Return.
	g.writeLn("return")

	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceStreamServer(s *parse.Stream, structName, srvcName string, errs []*parse.Error) {
	// Method declaration.
	g.writeLn("func (%s *%s) %s(s *orbit.Session, stream net.Conn) (err error) {", recv, structName, s.NamePrv())
	g.writeLn("defer stream.Close()")

	handlerArgs := "s"

	if !s.HasArgs() && !s.HasRet() {
		handlerArgs += ", stream"
	} else if s.HasArgs() {
		handlerArgs += ", args"

		g.writeLn("args := new%sReadChan(%s.s.CloserOneWay())", s.Args().Name, recv)
		g.writeLn("go func() {")
		g.writeLn("closingChan := args.ClosingChan()")
		g.writeLn("codec := %s.s.Codec()", recv)
		g.writeLn("for {")
		g.writeLn("arg := &%s{}", s.Args().Name)
		g.writeLn("err := packet.ReadDecode(stream, arg, codec)")
		g.writeLn("if err != nil {")
		g.writeLn("if args.IsClosing() { err = nil }")
		g.writeLn("args.setError(err)")
		g.writeLn("return")
		g.writeLn("}")
		g.writeLn("select {")
		g.writeLn("case <-closingChan:")
		g.writeLn("return")
		g.writeLn("case args.c <- arg:")
		g.writeLn("}")
		g.writeLn("}")
		g.writeLn("}()")
		g.writeLn("")
	}

	if s.HasRet() {
		handlerArgs += ", ret"

		g.writeLn("ret := new%sWriteChan(%s.s.CloserOneWay())", s.Ret().Name, recv)
		g.writeLn("go func() {")
		g.writeLn("closingChan := ret.ClosingChan()")
		g.writeLn("codec := %s.s.Codec()", recv)
		g.writeLn("for {")
		g.writeLn("select {")
		g.writeLn("case <- closingChan:")
		g.writeLn("return")
		g.writeLn("case data := <-ret.c:")
		g.writeLn("err := packet.WriteEncode(stream, data, codec)")
		g.writeLn("if err != nil {")
		g.writeLn("if ret.IsClosing() { err = nil }")
		g.writeLn("ret.setError(err)")
		g.writeLn("return")
		g.writeLn("}")
		g.writeLn("}")
		g.writeLn("}")
		g.writeLn("}()")
	}

	g.writeLn("err = %s.h.%s(%s)", recv, s.NamePub(), handlerArgs)
	g.errIfNil()
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}
