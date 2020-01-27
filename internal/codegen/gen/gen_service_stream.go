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

func (g *generator) genServiceStreamClient(s *ast.Stream, srvcName, structName string, errs []*ast.Error) {
	// Method declaration.
	write("func (%s *%s) ", recv, structName)
	write("%s(ctx context.Context) (", srvcName+s.Name)
	if s.Args != nil {
		write("args %sWriteChan, ", s.Args.String())
	}
	if s.Ret != nil {
		write("ret %sReadChan, ", s.Ret.String())
	} else if s.Args == nil {
		write("stream net.Conn, ")
	}
	write("err error)")
	writeLn(" {")

	// Method body.
	// First, open the stream.
	if s.Args == nil && s.Ret == nil {
		writeLn("return %s.s.OpenStream(ctx, Service%s, %s)", recv, srvcName, srvcName+s.Name)
		writeLn("}")
		writeLn("")
		return
	}

	writeLn("stream, err := %s.s.OpenStream(ctx, Service%s, %s)", recv, srvcName, srvcName+s.Name)
	errIfNil()

	if s.Args != nil {
		writeLn("args = new%sWriteChan(%s.s.CloserOneWay())", s.Args.Name, recv)
		writeLn("args.OnClosing(func() error { return stream.Close() })")
		writeLn("go func() {")
		writeLn("closingChan := args.ClosingChan()")
		writeLn("codec := %s.s.Codec()", recv)
		writeLn("for {")
		writeLn("select {")
		writeLn("case <- closingChan:")
		writeLn("return")
		writeLn("case arg := <-args.c:")
		writeLn("err := packet.WriteEncode(stream, arg, codec)")
		writeLn("if err != nil {")
		writeLn("if args.IsClosing() { err = nil }")
		writeLn("args.setError(err)")
		writeLn("return")
		writeLn("}")
		writeLn("}")
		writeLn("}")
		writeLn("}()")
	}

	if s.Ret != nil {
		writeLn("ret = new%sReadChan(%s.s.CloserOneWay())", s.Ret.Name, recv)
		writeLn("ret.OnClosing(func() error { return stream.Close() })")
		writeLn("go func() {")
		writeLn("closingChan := ret.ClosingChan()")
		writeLn("codec := %s.s.Codec()", recv)
		writeLn("for {")
		writeLn("data := &%s{}", s.Ret.Name)
		writeLn("err := packet.ReadDecode(stream, data, codec)")
		writeLn("if err != nil {")
		writeLn("if ret.IsClosing() { err = nil }")
		writeLn("ret.setError(err)")
		writeLn("return")
		writeLn("}")
		writeLn("select {")
		writeLn("case <-closingChan:")
		writeLn("return")
		writeLn("case ret.c <- data:")
		writeLn("}")
		writeLn("}")
		writeLn("}()")
	}

	// Return.
	writeLn("return")

	writeLn("}")
	writeLn("")
}

func (g *generator) genServiceStreamServer(s *ast.Stream, srvcName, srvcNamePrv, structName string, errs []*ast.Error) {
	// Method declaration.
	writeLn(
		"func (%s *%s) %s(s *orbit.Session, stream net.Conn) (err error) {",
		recv, structName, srvcNamePrv+s.Name,
	)
	writeLn("defer stream.Close()")

	handlerArgs := "s"

	if s.Args == nil && s.Ret == nil {
		handlerArgs += ", stream"
	} else if s.Args != nil {
		handlerArgs += ", args"

		writeLn("args := new%sReadChan(%s.s.CloserOneWay())", s.Args.Name, recv)
		writeLn("go func() {")
		writeLn("closingChan := args.ClosingChan()")
		writeLn("codec := %s.s.Codec()", recv)
		writeLn("for {")
		writeLn("arg := &%s{}", s.Args.Name)
		writeLn("err := packet.ReadDecode(stream, arg, codec)")
		writeLn("if err != nil {")
		writeLn("if args.IsClosing() { err = nil }")
		writeLn("args.setError(err)")
		writeLn("return")
		writeLn("}")
		writeLn("select {")
		writeLn("case <-closingChan:")
		writeLn("return")
		writeLn("case args.c <- arg:")
		writeLn("}")
		writeLn("}")
		writeLn("}()")
		writeLn("")
	}

	if s.Ret != nil {
		handlerArgs += ", ret"

		writeLn("ret := new%sWriteChan(%s.s.CloserOneWay())", s.Ret.Name, recv)
		writeLn("go func() {")
		writeLn("closingChan := ret.ClosingChan()")
		writeLn("codec := %s.s.Codec()", recv)
		writeLn("for {")
		writeLn("select {")
		writeLn("case <- closingChan:")
		writeLn("return")
		writeLn("case data := <-ret.c:")
		writeLn("err := packet.WriteEncode(stream, data, codec)")
		writeLn("if err != nil {")
		writeLn("if ret.IsClosing() { err = nil }")
		writeLn("ret.setError(err)")
		writeLn("return")
		writeLn("}")
		writeLn("}")
		writeLn("}")
		writeLn("}()")
	}

	writeLn("err = %s.h.%s(%s)", recv, srvcName+s.Name, handlerArgs)
	errIfNil()
	writeLn("return")
	writeLn("}")
	writeLn("")
}
