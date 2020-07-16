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

func (g *generator) genService(srvc *ast.Service, errs []*ast.Error) {
	// Create the call ids.
	g.writeLn("// CallIDs")
	g.writeLn("const (")
	for _, c := range srvc.Calls {
		g.writefLn("CallID%s = \"%s\"", c.Name, c.Name)
	}
	g.writeLn("// StreamIDs")
	for _, s := range srvc.Streams {
		g.writefLn("StreamID%s = \"%s\"", s.Name, s.Name)
	}
	g.writeLn(")")
	g.writeLn("")

	// Create the interfaces.
	g.genServiceClientInterface(srvc.Calls, srvc.Streams)
	g.genServiceInterface()
	g.genServiceHandlerInterface(srvc.Calls, srvc.Streams)

	// Create the private structs implementing the interfaces.
	g.genServiceClientStruct(srvc.Calls, srvc.Streams, errs)
	g.genServiceStruct(srvc.Calls, srvc.Streams, errs)
}

func (g *generator) genServiceClientInterface(calls []*ast.Call, streams []*ast.Stream) {
	g.writeLn("type Client interface {")
	g.writeLn("closer.Closer")

	if len(calls) > 0 {
		g.writeLn("// Calls")
		for _, c := range calls {
			g.genServiceClientCallSignature(c)
			g.writeLn("")
		}
	}

	if len(streams) > 0 {
		g.writeLn("// Streams")
		for _, s := range streams {
			g.genServiceClientStreamSignature(s)
			g.writeLn("")
		}
	}

	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceInterface() {
	g.writeLn("type Service interface {")
	g.writeLn("closer.Closer")
	g.writeLn("Run() error")
	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceHandlerInterface(calls []*ast.Call, streams []*ast.Stream) {
	// Generate Handler.
	g.writeLn("type ServiceHandler interface {")

	if len(calls) > 0 {
		g.writeLn("// Calls")
		for _, rc := range calls {
			g.genServiceHandlerCallSignature(rc)
		}
	}

	if len(streams) > 0 {
		g.writeLn("// Streams")
		for _, rs := range streams {
			g.genServiceHandlerStreamSignature(rs)
		}
	}

	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceClientStruct(calls []*ast.Call, streams []*ast.Stream, errs []*ast.Error) {
	// Generate the struct definition.
	g.writeLn("type client struct {")
	g.writeLn("oclient.Client")
	g.writeLn("codec codec.Codec")
	g.writeLn("callTimeout time.Duration")
	g.writeLn("streamInitTimeout time.Duration")
	g.writeLn("maxArgSize int")
	g.writeLn("maxRetSize int")
	g.writeLn("}")
	g.writeLn("")

	// Generate the constructor.
	g.writeLn("func NewClient(opts *oclient.Options) (c Client, err error) {")
	g.writeLn("oc, err := oclient.New(opts)")
	g.errIfNil()
	g.writeLn("c = &client{Client: oc, codec: opts.Codec, callTimeout: opts.CallTimeout, streamInitTimeout: opts.StreamInitTimeout, " +
		"maxArgSize: opts.MaxArgSize, maxRetSize:opts.MaxRetSize}")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")

	// Generate the calls.
	for _, c := range calls {
		g.genServiceClientCall(c, errs)
	}

	// Generate the streams.
	for _, s := range streams {
		g.genServiceClientStream(s, errs)
	}
}

func (g *generator) genServiceStruct(calls []*ast.Call, streams []*ast.Stream, errs []*ast.Error) {
	// Generate the struct definition.
	g.writeLn("type service struct {")
	g.writeLn("oservice.Service")
	g.writeLn("h ServiceHandler")
	g.writeLn("codec codec.Codec")
	g.writeLn("maxArgSize int")
	g.writeLn("maxRetSize int")
	g.writeLn("}")
	g.writeLn("")

	// Generate the constructor.
	g.writeLn("func NewService(h ServiceHandler, opts *oservice.Options) (s Service, err error) {")
	g.writeLn("os, err := oservice.New(opts)")
	g.errIfNil()
	g.writeLn("srvc := &service{Service: os, h: h, codec: opts.Codec, maxArgSize: opts.MaxArgSize, maxRetSize:opts.MaxRetSize}")
	// Ensure usage of service.
	// See https://github.com/desertbit/orbit/issues/34
	g.writeLn("// Ensure usage.")
	g.writeLn("_ = srvc")
	for _, c := range calls {
		g.genServiceHandlerCallRegister(c)
	}
	for _, s := range streams {
		g.genServiceHandlerStreamRegister(s)
	}
	g.writeLn("s = os")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")

	// Generate the calls.
	for _, c := range calls {
		g.genServiceHandlerCall(c)
	}

	// Generate the streams.
	for _, s := range streams {
		g.genServiceHandlerStream(s)
	}
}
