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
	"strings"

	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/utils"
)

func (g *generator) genServices(services []*ast.Service, errs []*ast.Error) {
	// Sort the services in lexicographical order.
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	for _, srvc := range services {
		g.writeLn("// %s  ---------------------", srvc.Name)

		// Generate the service.
		g.writeLn("// Service")
		g.genService(srvc, errs)

		g.writeLn("// ---------------------\n")
		g.writeLn("")
	}
}

func (g *generator) genService(srvc *ast.Service, errs []*ast.Error) {

	var (
		calls      = make([]*ast.Call, 0, len(srvc.Calls))
		revCalls   = make([]*ast.Call, 0, len(srvc.Calls))
		streams    = make([]*ast.Stream, 0, len(srvc.Streams))
		revStreams = make([]*ast.Stream, 0, len(srvc.Streams))
	)

	// Sort the entries into the respective categories.
	// Also create the call ids.
	g.writeLn("const (")
	g.writeLn("Service%s = \"%s\"", srvc.Name, srvc.Name)
	for _, c := range srvc.Calls {
		g.writeLn("%s = \"%s\"", srvc.Name+c.Name, c.Name)
		if c.Rev {
			revCalls = append(revCalls, c)
		} else {
			calls = append(calls, c)
		}
	}
	for _, s := range srvc.Streams {
		g.writeLn("%s = \"%s\"", srvc.Name+s.Name, s.Name)
		if s.Rev {
			revStreams = append(revStreams, s)
		} else {
			streams = append(streams, s)
		}
	}
	g.writeLn(")")
	g.writeLn("")

	// Create the interfaces.
	g.genServiceInterface("Consumer", srvc.Name, calls, revCalls, streams, revStreams)
	g.genServiceInterface("Provider", srvc.Name, revCalls, calls, revStreams, streams)

	// Create the private structs implementing the caller interfaces and providing the orbit handlers.
	g.genServiceStruct("Consumer", srvc.Name, calls, revCalls, streams, revStreams, errs)
	g.genServiceStruct("Provider", srvc.Name, revCalls, calls, revStreams, streams, errs)
}

func (g *generator) genServiceInterface(name, srvcName string, calls, revCalls []*ast.Call, streams, revStreams []*ast.Stream) {
	// Generate Caller.
	g.writeLn("type %s%sCaller interface {", srvcName, name)

	if len(calls) > 0 {
		g.writeLn("// Calls")
		for _, c := range calls {
			g.genServiceCallCallerSignature(c)
			g.writeLn("")
		}
	}

	if len(streams) > 0 {
		g.writeLn("// Streams")
		for _, s := range streams {
			g.genServiceStreamCallerSignature(s)
			g.writeLn("")
		}
	}

	g.writeLn("}")
	g.writeLn("")

	// Generate Handler.
	g.writeLn("type %s%sHandler interface {", srvcName, name)

	if len(revCalls) > 0 {
		g.writeLn("// Calls")
		for _, rc := range revCalls {
			g.genServiceCallHandlerSignature(rc)
			g.writeLn("")
		}
	}

	if len(revStreams) > 0 {
		g.writeLn("// Streams")
		for _, rs := range revStreams {
			g.genServiceStreamHandlerSignature(rs)
			g.writeLn("")
		}
	}

	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceStruct(
	name, srvcName string,
	calls, revCalls []*ast.Call,
	streams, revStreams []*ast.Stream,
	errs []*ast.Error,
) {
	srvcNamePrv := utils.ToLowerFirst(srvcName)

	// Write struct.
	strName := srvcNamePrv + name
	g.writeLn("type %s struct {", strName)
	g.writeLn("h %sHandler", srvcName+name)
	g.writeLn("s *orbit.Session")
	g.writeLn("}")
	g.writeLn("")

	// Generate constructor.
	g.genServiceStructConstructor(srvcName, srvcNamePrv, strName, revCalls, revStreams)

	// Generate the calls.
	for _, c := range calls {
		g.genServiceCallCaller(c, srvcName, strName, errs)
	}

	// Generate the rev calls.
	for _, rc := range revCalls {
		g.genServiceCallHandler(rc, strName, errs)
	}

	// Generate the streams.
	for _, s := range streams {
		g.genServiceStreamCaller(s, srvcName, strName, errs)
	}

	// generate the rev streams.
	for _, rs := range revStreams {
		g.genServiceStreamHandler(rs, strName, errs)
	}
}

func (g *generator) genServiceStructConstructor(srvcName, srvcNamePrv, name string, revCalls []*ast.Call, revStreams []*ast.Stream) {
	nameUp := strings.Title(name)
	g.writeLn("func Register%s(s *orbit.Session, h %sHandler) %sCaller {", nameUp, nameUp, nameUp)
	g.writeLn("cc := &%s{h: h, s: s}", name)
	for _, rc := range revCalls {
		g.writeLn("s.RegisterCall(Service%s, %s, cc.%s)", srvcName, srvcName+rc.Name, rc.NamePrv())
	}
	for _, rs := range revStreams {
		g.writeLn("s.RegisterStream(Service%s, %s, cc.%s)", srvcName, srvcName+rs.Name, rs.NamePrv())
	}
	g.writeLn("return cc")
	g.writeLn("}")
	g.writeLn("")
}
