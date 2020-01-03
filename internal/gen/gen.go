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
	"fmt"
	"io/ioutil"
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
)

func Generate(filePath string) (err error) {
	// Read the file.
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	// Parse the file data.
	services, types, err := parse.Parse(string(data))
	if err != nil {
		err = fmt.Errorf("could not parse %s: %v", filePath, err)
		return
	}

	// Write the preamble.
	dir := filepath.Dir(filePath)
	s := strings.Builder{} // Errors from write methods can be safely ignored.
	s.WriteString(fmt.Sprintf("/* code generated by orbit */\npackage %s\n\n", filepath.Base(dir)))

	// Sort the types in alphabetical order.
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})

	// Generate the type definitions.
	// Use a new builder, as we must first write the imports.
	s2 := strings.Builder{}
	imports := make(map[string]struct{})
	for _, t := range types {
		s2.WriteString("type " + t.Name + " ")
		genType(&s2, imports, t)
		s2.WriteString("\n\n")
	}

	// Generate the service definitions.
	// Use a new builder, as we must first write the imports.
	s3 := strings.Builder{}
	for _, srvc := range services {
		genService(&s3, imports, srvc)
	}

	// Write the imports.
	if len(imports) > 0 {
		s.WriteString("import (\n")
		for i := range imports {
			s.WriteString("\t\"" + i + "\"\n")
		}
		s.WriteString(")\n\n")
	}

	// Write the type definitions.
	s.WriteString("//#############//\n//### Types ###//\n//#############//\n\n")
	s.WriteString(s2.String())

	s.WriteString("//################//\n//### Services ###//\n//################//\n\n")
	s.WriteString(s3.String())

	// Write the contents to the file.
	return ioutil.WriteFile(
		filepath.Join(dir, strings.TrimSuffix(filepath.Base(filePath), OrbitSuffix)+genSuffix),
		[]byte(s.String()),
		filePerm,
	)
}

func genType(s *strings.Builder, imports map[string]struct{}, t parse.Type) {
	switch v := t.(type) {
	case *parse.StructType:
		// Sort its fields in alphabetical order.
		sort.Slice(v.Fields, func(i, j int) bool {
			return v.Fields[i].Name < v.Fields[j].Name
		})

		s.WriteString("struct {\n")
		for _, f := range v.Fields {
			s.WriteString(fmt.Sprintf("\t%s ", f.Name))

			// Structs just require a reference.
			if st, ok := f.Type.(*parse.StructType); ok {
				s.WriteString("*" + st.Name + "\n")
			} else {
				genType(s, imports, f.Type)
				s.WriteString("\n")
			}
		}
		s.WriteString("}")
	case *parse.MapType:
		s.WriteString("map[")
		// Key type can not be a struct.
		genType(s, imports, v.Key)
		s.WriteString("]")
		// Structs just require a reference.
		if st, ok := v.Value.(*parse.StructType); ok {
			s.WriteString("*" + st.Name)
		} else {
			genType(s, imports, v.Value)
		}
	case *parse.ArrType:
		s.WriteString("[]")
		// Structs just require a reference.
		if st, ok := v.ElemType.(*parse.StructType); ok {
			s.WriteString("*" + st.Name)
		} else {
			genType(s, imports, v.ElemType)
		}
	case *parse.BaseType:
		dt := v.DataType()

		// Check, if an import is needed.
		if dt == parse.TypeTime {
			if _, ok := imports["time"]; !ok {
				imports["time"] = struct{}{}
			}
			s.WriteString("time.Time")
		} else {
			s.WriteString(dt)
		}
	}
}

func genService(s *strings.Builder, imports map[string]struct{}, srvc *parse.Service) {
	s.WriteString("// " + srvc.Name + " ---------------------\n")

	var (
		calls      = make([]*parse.Call, 0)
		revCalls   = make([]*parse.Call, 0)
		streams    = make([]*parse.Stream, 0)
		revStreams = make([]*parse.Stream, 0)
	)

	// Sort the entries into the respective categories.
	for _, e := range srvc.Entries {
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

	// Imports!
	if len(streams) > 0 || len(revStreams) > 0 {
		imports["net"] = struct{}{}
	}

	// Create the interfaces.
	genCallerInterface(s, srvc.Name+"Consumer", calls, streams)
	genHandlerInterface(s, srvc.Name+"Consumer", revCalls, revStreams)
	genCallerInterface(s, srvc.Name+"Provider", revCalls, revStreams)
	genHandlerInterface(s, srvc.Name+"Provider", calls, streams)

	// Create the private structs implementing the caller interfaces.
	genCallerStruct(s, srvc.Name+"Consumer", calls, streams)
	genCallerStruct(s, srvc.Name+"Provider", revCalls, revStreams)

	s.WriteString("// ---------------------\n\n")
	return
}

func genCallerInterface(s *strings.Builder, name string, calls []*parse.Call, streams []*parse.Stream) {
	s.WriteString(fmt.Sprintf("type %sCaller interface {\n", name))
	s.WriteString("\t// Calls\n")
	for _, c := range calls {
		s.WriteString("\t" + c.Name() + "(")
		if c.Args != nil {
			s.WriteString("args *" + c.Args.Type.Name)
		}
		s.WriteString(")")
		s.WriteString(" (")
		if c.Ret != nil {
			s.WriteString("ret *" + c.Ret.Type.Name + ", ")
		}
		s.WriteString("err error)\n")
	}
	s.WriteString("\t// Streams\n")
	for _, st := range streams {
		s.WriteString(fmt.Sprintf("\t%s(conn net.Conn) (err error)\n", st.Name()))
	}
	s.WriteString("}\n\n")
}

func genHandlerInterface(s *strings.Builder, name string, calls []*parse.Call, streams []*parse.Stream) {
	s.WriteString(fmt.Sprintf("type %sHandler interface {\n", name))
	s.WriteString("\t// Calls\n")
	for _, c := range calls {
		s.WriteString("\t" + c.Name() + "(")
		if c.Args != nil {
			s.WriteString("args *" + c.Args.Type.Name)
		}
		s.WriteString(")")
		s.WriteString(" (")
		if c.Ret != nil {
			s.WriteString("ret *" + c.Ret.Type.Name + ", ")
		}
		s.WriteString("err error)\n")
	}
	s.WriteString("\t// Streams\n")
	for _, st := range streams {
		s.WriteString(fmt.Sprintf("\t%s(conn net.Conn) (err error)\n", st.Name()))
	}
	s.WriteString("}\n\n")
}

func genCallerStruct(s *strings.Builder, name string, calls []*parse.Call, streams []*parse.Stream) {
	// Write struct.
	strName := utils.ToLowerFirst(name) + "Caller"
	s.WriteString("type " + strName + " struct {\n")
	s.WriteString("\th " + name + "Handler\n")
	s.WriteString("}\n\n")

	// Implement Caller interface.
	const recv = "v1"
	for _, c := range calls {
		s.WriteString("func (" + recv + " *" + strName + ") " + c.Name())
		if c.Args != nil {
			s.WriteString("(args *" + c.Args.Type.Name + ")")
		}
		s.WriteString(" (")
		if c.Ret != nil {
			s.WriteString("ret *" + c.Ret.Type.Name + ", ")
		}
		s.WriteString("err error) {\n")
		// TODO:
		s.WriteString("}\n\n")
	}
}
