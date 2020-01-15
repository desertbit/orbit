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
	"strings"

	"github.com/desertbit/orbit/internal/parse"
	"github.com/desertbit/orbit/internal/utils"
)

func (g *generator) genServices(services []*parse.Service, errs []*parse.Error) {
	g.writeLn("//################//")
	g.writeLn("//### Services ###//")
	g.writeLn("//################//")
	g.writeLn("")

	for _, srvc := range services {
		g.genService(srvc, errs)
	}
}

func (g *generator) genService(srvc *parse.Service, errs []*parse.Error) {
	g.writeLn("// %s  ---------------------", srvc.Name)

	var (
		calls      = make([]*parse.Call, 0)
		revCalls   = make([]*parse.Call, 0)
		streams    = make([]*parse.Stream, 0)
		revStreams = make([]*parse.Stream, 0)
	)

	// Sort the entries into the respective categories.
	// Also create the call ids.
	g.writeLn("const (")
	g.writeLn("%s = \"%s\"", srvc.Name, srvc.Name)
	for _, e := range srvc.Entries {
		g.writeLn("%s = \"%s\"", srvc.Name+e.NamePub(), e.NamePub())
		switch v := e.(type) {
		case *parse.Call:
			if v.Rev() {
				revCalls = append(revCalls, v)
			} else {
				calls = append(calls, v)
			}
		case *parse.Stream:
			if v.Rev() {
				revStreams = append(revStreams, v)
			} else {
				streams = append(streams, v)
			}
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

	g.writeLn("// ---------------------\n")
	g.writeLn("")
	return
}

func (g *generator) genServiceInterface(name, srvcName string, calls, revCalls []*parse.Call, streams, revStreams []*parse.Stream) {
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
			g.write("%s(s *orbit.Session, ", rs.NamePub())
			if rs.HasArgs() {
				g.write("args %sReadChan, ", rs.Args().String())
			}
			if rs.HasRet() {
				g.write("ret %sWriteChan", rs.Ret().String())
			} else if !rs.HasArgs() {
				g.write("stream net.Conn")
			}
			g.write(") (err error)")
			g.writeLn("")
		}
	}

	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceStruct(
	name, srvcName string,
	calls, revCalls []*parse.Call,
	streams, revStreams []*parse.Stream,
	errs []*parse.Error,
) {
	// Write struct.
	strName := utils.ToLowerFirst(srvcName + name)
	g.writeLn("type %s struct {", strName)
	g.writeLn("h %sHandler", srvcName+name)
	g.writeLn("s *orbit.Session")
	g.writeLn("}")
	g.writeLn("")

	// Generate constructor.
	g.genServiceStructConstructor(strName, srvcName, revCalls, revStreams)

	// Generate the calls.
	for _, c := range calls {
		g.genServiceCallClient(c, strName, srvcName, errs)
	}

	// Generate the rev calls.
	for _, rc := range revCalls {
		g.genServiceCallServer(rc, strName, errs)
	}

	// Generate the streams.
	for _, s := range streams {
		g.genServiceStreamClient(s, strName, srvcName, errs)
	}

	// generate the rev streams.
	for _, rs := range revStreams {
		g.genServiceStreamServer(rs, strName, srvcName, errs)
	}
}

func (g *generator) genServiceStructConstructor(name, srvcName string, revCalls []*parse.Call, revStreams []*parse.Stream) {
	nameUp := strings.Title(name)
	g.writeLn("func Register%s(s *orbit.Session, h %sHandler) %sCaller {", nameUp, nameUp, nameUp)
	g.writeLn("cc := &%s{h: h, s: s}", name)
	for _, rc := range revCalls {
		g.writeLn("s.RegisterCall(%s, %s, cc.%s)", srvcName, srvcName+rc.NamePub(), rc.NamePrv())
	}
	for _, rs := range revStreams {
		g.writeLn("s.RegisterStream(%s, %s, cc.%s)", srvcName, srvcName+rs.NamePub(), rs.NamePrv())
	}
	g.writeLn("return cc")
	g.writeLn("}")
	g.writeLn("")
}
