/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2019 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2019 Sebastian Borchers <sebastian[at]desertbit.com>
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
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/desertbit/orbit/internal/parse"
	"github.com/desertbit/orbit/internal/utils"
)

const (
	OrbitSuffix = ".orbit"

	filePerm  = 0644
	genSuffix = "_gen.go"
	recv      = "v1"
)

func Generate(filePath string, streamChanSize uint) (err error) {
	// Read the file.
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	// Parse the file data.
	errs, services, types, err := parse.Parse(string(data))
	if err != nil {
		err = fmt.Errorf("could not parse %s\n-> %v", filePath, err)
		return
	}

	// Create generator.
	g := newGenerator()

	// Generate the errors.
	g.genErrors(errs)

	// Generate the type definitions.
	g.genTypes(types, services, streamChanSize)

	// Generate the service definitions.
	g.genServices(services, errs)

	// Write the preamble.
	var code string
	dir := filepath.Dir(filePath)
	code += fmt.Sprintf("/* code generated by orbit */\npackage %s\n\n", filepath.Base(dir))

	// Write the imports.
	if len(g.imps) > 0 {
		code += "import (\n"
		for imp := range g.imps {
			code += "\t\"" + imp + "\"\n"
		}
		code += ")\n\n"
	}

	// Write the contents to the file.
	goFilePath := filepath.Join(dir, strings.TrimSuffix(filepath.Base(filePath), OrbitSuffix)+genSuffix)
	err = ioutil.WriteFile(goFilePath, []byte(code+g.s.String()), filePerm)
	if err != nil {
		return
	}

	// Exec goimports, if available.
	err = execCmdIfExists("goimports", "-w", goFilePath)
	if err != nil {
		return
	}

	// Exec gofmt, if available.
	return execCmdIfExists("gofmt", "-s", "-w", goFilePath)
}

func execCmdIfExists(name string, args ...string) (err error) {
	cmd := exec.Command(name, args...)
	err = cmd.Run()
	if err != nil {
		var eErr *exec.ExitError
		if errors.Is(err, exec.ErrNotFound) {
			err = nil
		} else if errors.As(err, &eErr) {
			err = fmt.Errorf("%s: %v", name, string(eErr.Stderr))
		}
		return
	}
	return
}

type generator struct {
	s    strings.Builder
	imps imports
}

func newGenerator() *generator {
	return &generator{
		imps: newImports(),
	}
}

func (g *generator) genErrors(errs []*parse.Error) {
	g.writeLn("//##############//")
	g.writeLn("//### Errors ###//")
	g.writeLn("//##############//")
	g.writeLn("")

	if len(errs) == 0 {
		return
	}

	// Add import.
	g.imps.Add("errors")

	// Write error codes.
	g.writeLn("const (")
	for _, e := range errs {
		g.writeLn("ErrCode%s = %d", e.Name, e.ID)
	}
	g.writeLn(")")

	// Write standard error variables along with the orbit control.ErrorCode ones.
	g.writeLn("var (")
	for _, e := range errs {
		g.writeLn("Err%s = errors.New(\"%s\")", e.Name, strExplode(e.Name))
		g.writeLn("orbitErr%s = orbit.Err(Err%s, Err%s.Error(), ErrCode%s)", e.Name, e.Name, e.Name, e.Name)
	}
	g.writeLn(")")
}

func (g *generator) genTypes(ts []*parse.StructType, srvcs []*parse.Service, streamChanSize uint) {
	g.writeLn("//#############//")
	g.writeLn("//### Types ###//")
	g.writeLn("//#############//")
	g.writeLn("")

	// Sort the types in alphabetical order.
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Name < ts[j].Name
	})

NextType:
	for _, t := range ts {
		// Sort its fields in alphabetical order.
		sort.Slice(t.Fields, func(i, j int) bool {
			return t.Fields[i].Name < t.Fields[j].Name
		})

		g.writeLn("type %s struct {", t.Name)
		for _, f := range t.Fields {
			g.write("%s ", f.Name)
			g.genType(f.Type)
			g.writeLn("")
		}
		g.writeLn("}")
		g.writeLn("")

		// Generate a chan type, if it is used in a stream as arg or ret value.
		var generatedArgs, generatedRet bool
		for _, srvc := range srvcs {
			for _, e := range srvc.Entries {
				if s, ok := e.(*parse.Stream); ok {
					if s.Args != nil && s.Args.Type == t && !generatedArgs {
						g.genChanType(t.Name, false, streamChanSize)
						generatedArgs = true
					}
					if s.Ret != nil && s.Ret.Type == t && !generatedRet {
						g.genChanType(t.Name, true, streamChanSize)
						generatedRet = true
					}

					if generatedArgs && generatedRet {
						continue NextType
					}
				}
			}
		}
	}
}

func (g *generator) genType(t parse.Type) {
	switch v := t.(type) {
	case *parse.StructType:
		// Structs just require a reference.
		g.write("*%s", v.Name)
	case *parse.MapType:
		g.write("map[")
		// Generate Key type.
		g.genType(v.Key)
		g.write("]")
		// Generate Value type.
		g.genType(v.Value)
	case *parse.ArrType:
		g.write("[]")
		// Generate Elem type.
		g.genType(v.ElemType)
	case *parse.BaseType:
		dt := v.DataType()

		// Check, if an import is needed.
		if dt == parse.TypeTime {
			g.imps.Add("time")
			g.write("time.Time")
		} else {
			g.write(dt)
		}
	}
}

func (g *generator) genChanType(name string, ro bool, streamChanSize uint) {
	suffix := "Args"
	if ro {
		suffix = "Ret"
	}

	g.imps.Add("sync")
	g.imps.Add("github.com/desertbit/closer/v3")

	g.writeLn("type %sChan%s struct {", name, suffix)
	g.writeLn("closer.Closer")
	g.write("C ")
	if ro {
		g.write("<-chan ")
	} else {
		g.write("chan<- ")
	}
	g.writeLn("*%s", name)
	g.writeLn("c chan *%s", name)
	g.writeLn("mx sync.Mutex")
	g.writeLn("err error")
	g.writeLn("}")
	g.writeLn("")

	g.writeLn("func new%sChan%s(cl closer.Closer) *%sChan%s {", name, suffix, name, suffix)
	g.writeLn("c := &%sChan%s{Closer: cl, c: make(chan *%s, %d)}", name, suffix, name, streamChanSize)
	g.writeLn("c.C = c.c")
	g.writeLn("return c")
	g.writeLn("}")
	g.writeLn("")

	g.writeLn("func (c *%sChan%s) setError(err error) {", name, suffix)
	g.writeLn("c.mx.Lock()")
	g.writeLn("c.err = err")
	g.writeLn("c.mx.Unlock()")
	g.writeLn("c.Close_()")
	g.writeLn("}")
	g.writeLn("")

	g.writeLn("func (c *%sChan%s) Err() (err error) {", name, suffix)
	g.writeLn("c.mx.Lock()")
	g.writeLn("err = c.err")
	g.writeLn("c.mx.Unlock()")
	g.writeLn("return")
	g.writeLn("}")
	g.writeLn("")
}

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
	for _, e := range srvc.Entries {
		g.writeLn("%s = \"%s\"", srvc.Name+e.Name(), e.Name())
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

	// Imports!
	if len(srvc.Entries) > 0 {
		g.imps.Add("github.com/desertbit/orbit/pkg/orbit")
	}
	if len(streams) > 0 || len(revStreams) > 0 {
		g.imps.Add("net")
	}

	// Create the interfaces.
	g.genServiceInterface("ConsumerCaller", srvc.Name, calls, streams)
	g.genServiceInterface("ConsumerHandler", srvc.Name, revCalls, revStreams)
	g.genServiceInterface("ProviderCaller", srvc.Name, revCalls, revStreams)
	g.genServiceInterface("ProviderHandler", srvc.Name, calls, streams)
	g.writeLn("")

	// Create the private structs implementing the caller interfaces and providing the orbit handlers.
	g.genServiceStruct("Consumer", srvc.Name, calls, revCalls, streams, revStreams, errs)
	g.genServiceStruct("Provider", srvc.Name, revCalls, calls, revStreams, streams, errs)

	g.writeLn("// ---------------------\n")
	g.writeLn("")
	return
}

func (g *generator) genServiceInterface(name, srvcName string, calls []*parse.Call, streams []*parse.Stream) {
	g.writeLn("type %s interface {", srvcName+name)
	g.writeLn("// Calls")
	for _, c := range calls {
		g.genServiceCallSignature(c)
		g.writeLn("")
	}
	g.writeLn("// Streams")
	for _, s := range streams {
		g.genServiceStreamSignature(s)
		g.writeLn("")
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
	g.genServiceStructConstructor(strName, srvcName)

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

func (g *generator) genServiceStructConstructor(name, srvcName string) {
	nameUp := strings.Title(name)
	g.writeLn("func Register%s(s *orbit.Session, h %sHandler) %sCaller {", nameUp, nameUp, nameUp)
	g.writeLn("cc := &%s{h: h, s: s}", name)
	g.writeLn("return cc")
	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceCallClient(c *parse.Call, structName, srvcName string, errs []*parse.Error) {
	// Method declaration.
	g.write("func (%s *%s) ", recv, structName)
	g.genServiceCallSignature(c)
	g.writeLn(" {")

	// Method body.
	// First, make the call.
	if c.Ret != nil {
		g.write("retData, err := ")
	} else {
		g.write("_, err = ")
	}
	g.write("%s.s.Call(ctx, %s, ", recv, srvcName+c.Name())
	if c.Args != nil {
		g.writeLn("args)")
	} else {
		g.writeLn("nil)")
	}

	// Check error and parse control.ErrorCodes.
	g.writeErrCheckOrbitCaller(errs)

	// If return arguments are expected, decode them.
	if c.Ret != nil {
		g.writeLn("err = retData.Decode(ret)")
		g.writeErrCheck()
	}

	// Return.
	g.writeLn("return")

	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceCallServer(c *parse.Call, structName string, errs []*parse.Error) {
	// Method declaration.
	g.writeLn(
		"func (%s *%s) %s(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {",
		recv, structName, utils.ToLowerFirst(c.Name()),
	)

	// Method body.
	// Parse the args.
	handlerArgs := "ctx"
	if c.Args != nil {
		handlerArgs += ", args"
		g.writeLn("var args *%s", c.Args.Type.Name)
		g.writeLn("err = ad.Decode(args)")
		g.writeErrCheck()
	}

	// Call the handler.
	if c.Ret != nil {
		g.writeLn("ret, err := %s.h.%s(%s)", recv, c.Name(), handlerArgs)
	} else {
		g.writeLn("err = %s.h.%s(%s)", recv, c.Name(), handlerArgs)
	}

	// Check error and convert to orbit errors.
	g.writeErrCheckOrbitHandler(errs)

	// Assign return value.
	if c.Ret != nil {
		g.writeLn("r = ret")
	}

	// Return.
	g.writeLn("return")

	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceCallSignature(c *parse.Call) {
	g.imps.Add("context")

	g.write("%s(ctx context.Context", c.Name())
	if c.Args != nil {
		g.write(", args *%s", c.Args.Type.Name)
	}
	g.write(") (")
	if c.Ret != nil {
		g.write("ret *%s, ", c.Ret.Type.Name)
	}
	g.write("err error)")
}

func (g *generator) genServiceStreamClient(s *parse.Stream, structName, srvcName string, errs []*parse.Error) {
	// Method declaration.
	g.write("func (%s *%s) ", recv, structName)
	g.genServiceStreamSignature(s)
	g.writeLn(" {")

	// Method body.
	// First, open the stream.
	if s.Args == nil && s.Ret == nil {
		g.writeLn("return %s.s.OpenStream(ctx, %s)", recv, srvcName+s.Name())
		g.writeLn("}")
		g.writeLn("")
		return
	}

	g.writeLn("conn, err := %s.s.OpenStream(ctx, %s)", recv, srvcName+s.Name())
	g.writeErrCheck()

	if s.Args != nil {
		g.imps.Add("github.com/desertbit/orbit/internal/packet")

		g.writeLn("args = new%sChanArgs(%s.s.CloserOneWay())", s.Args.Type.Name, recv)
		g.writeLn("go func() {")
		g.writeLn("closingChan := args.ClosingChan()")
		g.writeLn("for {")
		g.writeLn("select {")
		g.writeLn("case <- closingChan:")
		g.writeLn("return")
		g.writeLn("case arg := <-args.c:")
		g.writeLn("err := packet.WriteEncode(conn, arg, %s.s.Codec())", recv)
		g.writeLn("if err != nil && !%s.s.IsClosing() {", recv)
		g.writeLn("args.setError(err)")
		g.writeLn("return")
		g.writeLn("}")
		g.writeLn("}")
		g.writeLn("}")
		g.writeLn("}()")
	}

	if s.Ret != nil {
		g.imps.Add("github.com/desertbit/orbit/internal/packet")

		g.writeLn("ret = new%sChanRet(%s.s.CloserOneWay())", s.Ret.Type.Name, recv)
		g.writeLn("go func() {")
		g.writeLn("closingChan := ret.ClosingChan()")
		g.writeLn("for {")
		g.writeLn("select {")
		g.writeLn("case <- closingChan:")
		g.writeLn("return")
		g.writeLn("default:")
		g.writeLn("var data *Plate")
		g.writeLn("err := packet.ReadDecode(conn, data, %s.s.Codec())", recv)
		g.writeLn("if err != nil && !%s.s.IsClosing() {", recv)
		g.writeLn("ret.setError(err)")
		g.writeLn("}")
		g.writeLn("select {")
		g.writeLn("case <-closingChan:")
		g.writeLn("return")
		g.writeLn("case ret.c <- data:")
		g.writeLn("}")
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
	g.writeLn("func (%s *%s) %s(s *orbit.Session, stream net.Conn) (err error) {", recv, structName, s.Name())
	g.writeLn("")
	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceStreamSignature(s *parse.Stream) {
	g.imps.Add("context")
	g.imps.Add("net")

	g.write("%s(s *orbit.Session, stream net.Conn) (", s.Name())
	if s.Args != nil {
		g.write("args *Chan%sArgs, ", s.Args.Type.Name)
	}
	if s.Ret != nil {
		g.write("ret *Chan%sRet, ", s.Ret.Type.Name)
	} else if s.Args == nil {
		g.write("conn net.Conn, ")
	}
	g.write("err error)")
}

func (g *generator) writeErrCheck() {
	g.writeLn("if err != nil { return }")
}

func (g *generator) writeErrCheckOrbitCaller(errs []*parse.Error) {
	g.writeLn("if err != nil {")
	// Check, if a control.ErrorCode has been returned.
	if len(errs) > 0 {
		g.imps.Add("errors")

		g.writeLn("var cErr *orbit.ErrorCode")
		g.writeLn("if errors.As(err, &cErr) {")
		g.writeLn("switch cErr.Code {")
		for _, e := range errs {
			g.writeLn("case %d:", e.ID)
			g.writeLn("err = Err%s", e.Name)
		}
		g.writeLn("}")
		g.writeLn("}")
	}
	g.writeLn("return")
	g.writeLn("}")
}

func (g *generator) writeErrCheckOrbitHandler(errs []*parse.Error) {
	g.writeLn("if err != nil {")
	// Check, if a api error has been returned and convert it to a control.ErrorCode.
	if len(errs) > 0 {
		g.imps.Add("errors")

		for i, e := range errs {
			g.writeLn("if errors.Is(err, Err%s) {", e.Name)
			g.writeLn("err = orbitErr%s", e.Name)
			if i < len(errs)-1 {
				g.write("} else ")
			} else {
				g.writeLn("}")
			}
		}
	}
	g.writeLn("return")
	g.writeLn("}")
}

func (g *generator) writeLn(format string, a ...interface{}) {
	g.write(format, a...)
	g.s.WriteString("\n")
}

func (g *generator) write(format string, a ...interface{}) {
	if len(a) == 0 {
		g.s.WriteString(format)
		return
	}

	g.s.WriteString(fmt.Sprintf(format, a...))
}
