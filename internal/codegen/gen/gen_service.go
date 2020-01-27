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

func (g *generator) genServices(services []*ast.Service, globErrs []*ast.Error, streamChanSize uint) {
	// Sort the services in lexicographical order.
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	for _, srvc := range services {
		writeLn("// %s  ---------------------", srvc.Name)

		// Generate the errors.
		writeLn("// Errors ")
		genErrors(srvc.Errors)

		// Generate the types.
		writeLn("// Types")
		genTypes(srvc.Types, []*ast.Service{srvc}, streamChanSize)

		// Generate the service.
		writeLn("// Service")
		g.genService(srvc, globErrs)

		writeLn("// ---------------------\n")
		writeLn("")
	}
}

func (g *generator) genService(srvc *ast.Service, globErrs []*ast.Error) {

	var (
		calls      = make([]*ast.Call, 0, len(srvc.Calls))
		revCalls   = make([]*ast.Call, 0, len(srvc.Calls))
		streams    = make([]*ast.Stream, 0, len(srvc.Streams))
		revStreams = make([]*ast.Stream, 0, len(srvc.Streams))
	)

	// Sort the entries into the respective categories.
	// Also create the call ids.
	writeLn("const (")
	writeLn("Service%s = \"%s\"", srvc.Name, srvc.Name)
	for _, c := range srvc.Calls {
		writeLn("%s = \"%s\"", srvc.Name+c.Name, c.Name)
		if c.Rev {
			revCalls = append(revCalls, c)
		} else {
			calls = append(calls, c)
		}
	}
	for _, s := range srvc.Streams {
		writeLn("%s = \"%s\"", srvc.Name+s.Name, s.Name)
		if s.Rev {
			revStreams = append(revStreams, s)
		} else {
			streams = append(streams, s)
		}
	}
	writeLn(")")
	writeLn("")

	// Create the interfaces.
	g.genServiceInterface("Consumer", srvc.Name, calls, revCalls, streams, revStreams)
	g.genServiceInterface("Provider", srvc.Name, revCalls, calls, revStreams, streams)

	// Create the private structs implementing the caller interfaces and providing the orbit handlers.
	// Handle not only the global errors, but the service's ones as well.
	errs := append(globErrs, srvc.Errors...)
	g.genServiceStruct("Consumer", srvc.Name, calls, revCalls, streams, revStreams, errs)
	g.genServiceStruct("Provider", srvc.Name, revCalls, calls, revStreams, streams, errs)
}

func (g *generator) genServiceInterface(name, srvcName string, calls, revCalls []*ast.Call, streams, revStreams []*ast.Stream) {
	// Generate Caller.
	writeLn("type %s%sCaller interface {", srvcName, name)

	if len(calls) > 0 {
		writeLn("// Calls")
		for _, c := range calls {
			genServiceCallCallerSignature(c, srvcName)
			writeLn("")
		}
	}

	if len(streams) > 0 {
		writeLn("// Streams")
		for _, s := range streams {
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
			writeLn("")
		}
	}

	writeLn("}")
	writeLn("")

	// Generate Handler.
	writeLn("type %s%sHandler interface {", srvcName, name)

	if len(revCalls) > 0 {
		writeLn("// Calls")
		for _, rc := range revCalls {
			genServiceCallHandlerSignature(rc, srvcName)
			writeLn("")
		}
	}

	if len(revStreams) > 0 {
		writeLn("// Streams")
		for _, rs := range revStreams {
			write("%s(s *orbit.Session, ", srvcName+rs.Name)
			if rs.Args != nil {
				write("args %sReadChan, ", rs.Args.String())
			}
			if rs.Ret != nil {
				write("ret %sWriteChan", rs.Ret.String())
			} else if rs.Args == nil {
				write("stream net.Conn")
			}
			write(") (err error)")
			writeLn("")
		}
	}

	writeLn("}")
	writeLn("")
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
	writeLn("type %s struct {", strName)
	writeLn("h %sHandler", srvcName+name)
	writeLn("s *orbit.Session")
	writeLn("}")
	writeLn("")

	// Generate constructor.
	g.genServiceStructConstructor(srvcName, srvcNamePrv, strName, revCalls, revStreams)

	// Generate the calls.
	for _, c := range calls {
		genServiceCallClient(c, srvcName, strName, errs)
	}

	// Generate the rev calls.
	for _, rc := range revCalls {
		genServiceCallServer(rc, srvcName, srvcNamePrv, strName, errs)
	}

	// Generate the streams.
	for _, s := range streams {
		genServiceStreamClient(s, srvcName, strName, errs)
	}

	// generate the rev streams.
	for _, rs := range revStreams {
		genServiceStreamServer(rs, srvcName, srvcNamePrv, strName, errs)
	}
}

func (g *generator) genServiceStructConstructor(srvcName, srvcNamePrv, name string, revCalls []*ast.Call, revStreams []*ast.Stream) {
	nameUp := strings.Title(name)
	writeLn("func Register%s(s *orbit.Session, h %sHandler) %sCaller {", nameUp, nameUp, nameUp)
	writeLn("cc := &%s{h: h, s: s}", name)
	for _, rc := range revCalls {
		writeLn("s.RegisterCall(Service%s, %s, cc.%s)", srvcName, srvcName+rc.Name, srvcNamePrv+rc.Name)
	}
	for _, rs := range revStreams {
		writeLn("s.RegisterStream(Service%s, %s, cc.%s)", srvcName, srvcName+rs.Name, srvcNamePrv+rs.Name)
	}
	writeLn("return cc")
	writeLn("}")
	writeLn("")
}
