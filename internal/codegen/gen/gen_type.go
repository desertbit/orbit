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

	// We want to send our orbit Errors over the wire with orbit-managed streams.
	// Therefore, we must build a struct containing both the error and the code.
	// But, only generate this struct if we have at least one such stream.
	var generatedErrStruct bool

	// Generate a stream type for every stream arg or ret or bidirectional stream func,
	// but only once!
	genStreams := make(map[string]struct{})
	for _, s := range srvc.Streams {
		if s.Arg == nil && s.Ret == nil {
			continue
		}

		if !generatedErrStruct {
			generatedErrStruct = true
			g.genStreamErrStruct()
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

func (g *generator) genStreamErrStruct() {
	// Generate a struct representing an orbit error.
	g.writefLn("type %s struct {", streamErrCode)
	g.writeLn("Err string")
	g.writeLn("Code int")
	g.writeLn("}")
}

// genClientReadStreamType generates the client read side of a stream.
func (g *generator) genClientReadStreamType(dt ast.DataType) {
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
	g.writeLn("cl.OnClosing(s.Close)")
	g.writefLn("return &%sReadStream{Closer: cl, stream: s, codec: cc, maxSize: ms}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writefLn("func (%s *%sReadStream) Read() (ret %s, err error) {", recv, dt.Name(), dt.Decl())
	g.writefLn("err = packet.ReadDecode(%s.stream, &ret, %s.codec, %s.maxSize)", recv, recv, recv)
	g.writeLn("if err != nil {")
	g.writefLn("if errors.Is(err, packet.ErrZeroData) || errors.Is(err, io.EOF) || %s.IsClosing() || %s.stream.IsClosed() {", recv, recv)
	g.writeLn("err = ErrClosed")
	g.writeLn("}")
	g.writefLn("%s.Close_()", recv)
	g.writeLn("return")
	g.writeLn("}")
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
	g.writeLn("closer.Closer")
	g.writeLn("stream transport.Stream")
	g.writeLn("codec codec.Codec")
	g.writeLn("maxSize int")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sWriteStream(cl closer.Closer, s transport.Stream, cc codec.Codec, ms int) *%sWriteStream {", dt.Name(), dt.Name())
	g.writeLn("cl.OnClosing(s.Close)")
	g.writeLn("cl.OnClosing(func() error { return packet.Write(s, nil, 0) })")
	g.writefLn("return &%sWriteStream{Closer: cl, stream: s, codec: cc, maxSize: ms}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writefLn("func (%s *%sWriteStream) Write(arg %s) (err error) {", recv, dt.Name(), dt.Decl())
	g.writefLn("err = packet.WriteEncode(%s.stream, &arg, %s.codec, %s.maxSize)", recv, recv, recv)
	g.writeLn("if err != nil {")
	g.writefLn("if errors.Is(err, io.EOF) || %s.IsClosing() || %s.stream.IsClosed() {", recv, recv)
	g.writefLn("%s.Close_()", recv)
	g.writeLn("return ErrClosed")
	g.writeLn("}")
	g.writeLn("}")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

// genClientStreamType generates the client side of a bidirectional stream.
func (g *generator) genClientStreamType(name string, arg, ret ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sClientStream", name)
	g.writefLn("type %sClientStream struct {", name)
	g.writeLn("closer.Closer")
	g.writeLn("stream transport.Stream")
	g.writeLn("codec codec.Codec")
	g.writeLn("maxSize int")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sClientStream(cl closer.Closer, s transport.Stream, cc codec.Codec, ms int) *%sClientStream {", name, name)
	g.writeLn("cl.OnClosing(s.Close)")
	g.writeLn("// Since we have a writing side, ensure to send the zero packet once we close.")
	g.writeLn("cl.OnClosing(func() error { return packet.Write(s, nil, 0) })")
	g.writefLn("return &%sClientStream{Closer: cl, stream: s, codec: cc, maxSize: ms}", name)
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writefLn("func (%s *%sClientStream) Read() (ret %s, err error) {", recv, name, ret.Name())
	g.writefLn("err = packet.ReadDecode(%s.stream, &ret, %s.codec, %s.maxSize)", recv, recv, recv)
	g.writeLn("if err != nil {")
	g.writefLn("if errors.Is(err, packet.ErrZeroData) || errors.Is(err, io.EOF) || %s.IsClosing() || %s.stream.IsClosed() {", recv, recv)
	g.writeLn("err = ErrClosed")
	g.writeLn("}")
	g.writefLn("%s.Close_()", recv)
	g.writeLn("return")
	g.writeLn("}")
	// Validate, if needed.
	g.writeValErrCheck(ret, "ret")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writefLn("func (%s *%sClientStream) Write(arg %s) (err error) {", recv, name, arg.Name())
	g.writefLn("err = packet.WriteEncode(%s.stream, &arg, %s.codec, %s.maxSize)", recv, recv, recv)
	g.writeLn("if err != nil {")
	g.writefLn("if errors.Is(err, io.EOF) || %s.IsClosing() || %s.stream.IsClosed() {", recv, recv)
	g.writefLn("%s.Close_()", recv)
	g.writeLn("return ErrClosed")
	g.writeLn("}")
	g.writeLn("}")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

// genServiceReadStreamType generates the service read side of a stream.
func (g *generator) genServiceReadStreamType(dt ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sReadStream", dt.Name())
	g.writefLn("type %sReadStream struct {", dt.Name())
	g.writeLn("stream transport.Stream")
	g.writeLn("codec codec.Codec")
	g.writeLn("maxSize int")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sReadStream(s transport.Stream, cc codec.Codec, ms int) *%sReadStream {", dt.Name(), dt.Name())
	g.writefLn("return &%sReadStream{stream: s, codec: cc, maxSize: ms}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writefLn("func (%s *%sReadStream) Read() (arg %s, err error) {", recv, dt.Name(), dt.Decl())
	g.writefLn("err = packet.ReadDecode(%s.stream, &arg, %s.codec, %s.maxSize)", recv, recv, recv)
	g.writeLn("if err != nil {")
	g.writefLn("if errors.Is(err, packet.ErrZeroData) || errors.Is(err, io.EOF) || %s.stream.IsClosed() {", recv)
	g.writeLn("err = ErrClosed")
	g.writeLn("}")
	g.writeLn("return")
	g.writeLn("}")
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
	g.writeLn("stream transport.Stream")
	g.writeLn("codec codec.Codec")
	g.writeLn("maxSize int")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sWriteStream(s transport.Stream, cc codec.Codec, ms int) *%sWriteStream {", dt.Name(), dt.Name())
	g.writefLn("return &%sWriteStream{stream: s, codec: cc, maxSize: ms}", dt.Name())
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writefLn("func (%s *%sWriteStream) Write(ret %s) (err error) {", recv, dt.Name(), dt.Decl())
	g.writefLn("err = packet.WriteEncode(%s.stream, &ret, %s.codec, %s.maxSize)", recv, recv, recv)
	g.writeLn("if err != nil {")
	g.writefLn("if errors.Is(err, io.EOF) || %s.stream.IsClosed() {", recv)
	g.writeLn("return ErrClosed")
	g.writeLn("}")
	g.writeLn("}")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

// genServiceStreamType generates the service side of a bidirectional stream.
func (g *generator) genServiceStreamType(name string, arg, ret ast.DataType) {
	// Type definition.
	g.writefLn("//msgp:ignore %sServiceStream", name)
	g.writefLn("type %sServiceStream struct {", name)
	g.writeLn("stream transport.Stream")
	g.writeLn("codec codec.Codec")
	g.writeLn("maxSize int")
	g.writeLn("}")
	g.writeLn("")

	// Constructor.
	g.writefLn("func new%sServiceStream(s transport.Stream, cc codec.Codec, ms int) *%sServiceStream {", name, name)
	g.writefLn("return &%sServiceStream{stream: s, codec: cc, maxSize: ms}", name)
	g.writeLn("}")
	g.writeLn("")

	// Read method.
	g.writefLn("func (%s *%sServiceStream) Read() (arg %s, err error) {", recv, name, arg.Decl())
	g.writefLn("err = packet.ReadDecode(%s.stream, &arg, %s.codec, %s.maxSize)", recv, recv, recv)
	g.writeLn("if err != nil {")
	g.writefLn("if errors.Is(err, packet.ErrZeroData) || errors.Is(err, io.EOF) || %s.stream.IsClosed() {", recv)
	g.writeLn("err = ErrClosed")
	g.writeLn("}")
	g.writeLn("return")
	g.writeLn("}")
	// Validate, if needed.
	g.writeValErrCheck(arg, "arg")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")

	// Write method.
	g.writefLn("func (%s *%sServiceStream) Write(ret %s) (err error) {", recv, name, ret.Decl())
	g.writefLn("err = packet.WriteEncode(%s.stream, &ret, %s.codec, %s.maxSize)", recv, recv, recv)
	g.writeLn("if err != nil {")
	g.writefLn("if errors.Is(err, io.EOF) || %s.stream.IsClosed() {", recv)
	g.writeLn("return ErrClosed")
	g.writeLn("}")
	g.writeLn("}")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}
