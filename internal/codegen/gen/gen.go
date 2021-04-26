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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/lexer"
	"github.com/desertbit/orbit/internal/codegen/parser"
	"github.com/desertbit/orbit/internal/codegen/validate"
	"github.com/rs/zerolog/log"
)

const (
	dirPerm  = 0755
	filePerm = 0666

	orbitSuffix       = ".orbit"
	genOrbitSuffix    = "_orbit_gen.go"
	genMsgpSuffix     = "_msgp_gen.go"
	genMsgpTestSuffix = "_msgp_gen_test.go"

	recv = "v1"
)

func Generate(orbitFile string, force bool) (err error) {
	// Check the file suffix.
	if !strings.HasSuffix(orbitFile, orbitSuffix) {
		return fmt.Errorf("'%s' is not an orbit file, missing '%s' suffix", orbitFile, orbitSuffix)
	}

	// Ensure, the file's path is absolute.
	orbitFile, err = filepath.Abs(orbitFile)
	if err != nil {
		return
	}

	// Check, if the file has been modified.
	modified, err := compareWithGenCache(orbitFile, force)
	if err != nil {
		if errors.Is(err, errCacheInvalid) {
			log.Warn().Err(err).Msg("invalid old cache, generating all files and overwriting cache")
		} else if !errors.Is(err, errCacheNotFound) {
			return
		}
		err = nil
	} else if !modified {
		return
	}

	// Read whole file content.
	input, err := ioutil.ReadFile(orbitFile)
	if err != nil {
		return
	}

	// Wrap a lexer around it.
	lx := lexer.Lex(string(input))

	// Parse the lexer output and create an AST.
	f, err := parser.Parse(lx)
	if err != nil {
		return
	}

	// Validate the produced AST.
	err = validate.Validate(f)
	if err != nil {
		return
	}

	// The name of the generated file is the same as the orbit file,
	// but with a different file ending.
	filePathNoSuffix := strings.TrimSuffix(orbitFile, orbitSuffix)
	ofp := filePathNoSuffix + genOrbitSuffix

	// Generate the code into a single file.
	pkgName := filepath.Base(filepath.Dir(orbitFile))
	err = ioutil.WriteFile(ofp, []byte(generate(pkgName, f)), filePerm)
	if err != nil {
		return
	}

	// Format the file and simplify the code, where possible.
	err = execCmd("gofmt", "-s", "-w", ofp)
	if err != nil {
		return
	}

	// Update the cache for this file.
	err = updateGenCache(orbitFile)
	if err != nil {
		return
	}

	// Generate msgp code for it, if at least one type has been defined.
	mfp := filePathNoSuffix + genMsgpSuffix
	if len(f.Types) > 0 {
		err = execCmd("msgp", "-file", ofp, "-o", mfp)
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				err = errors.New("msgp required to generate MessagePack code")
			}
			return
		}
	} else {
		// Otherwise, ensure our old msgp files (including test) are deleted.
		err = os.Remove(mfp)
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
			} else {
				return
			}
		}
		err = os.Remove(filePathNoSuffix + genMsgpTestSuffix)
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
			} else {
				return
			}
		}
	}
	return
}

type generator struct {
	s strings.Builder
}

func generate(pkgName string, f *ast.File) string {
	g := generator{}

	// Write the preamble.
	g.writeLn("/* code generated by orbit */")
	g.writefLn("package %s", pkgName)
	g.writeLn("")

	// Write the imports.
	imports := [][2]string{
		{"context", "context"},
		{"errors", "errors"},
		{"fmt", "fmt"},
		{"io", "io"},
		{"net", "net"},
		{"time", "time"},
		{"strings", "strings"},
		{"sync", "sync"},
		{"oclient", "github.com/desertbit/orbit/pkg/client"},
		{"closer", "github.com/desertbit/closer/v3"},
		{"codec", "github.com/desertbit/orbit/pkg/codec"},
		{"packet", "github.com/desertbit/orbit/pkg/packet"},
		{"oservice", "github.com/desertbit/orbit/pkg/service"},
		{"transport", "github.com/desertbit/orbit/pkg/transport"},
		{"validator", "github.com/go-playground/validator/v10"},
	}

	g.writeLn("import (")
	for _, imp := range imports {
		g.writefLn(`%s "%s"`, imp[0], imp[1])
	}
	g.writeLn(")")
	g.writeLn("")

	// Generate a var block that imports one struct from every package to ensure usage.
	g.writeLn("// Ensure that all imports are used.")
	g.writeLn("var (")
	g.writeLn("_ context.Context")
	g.writeLn("_ = errors.New(\"\")")
	g.writeLn("_ = fmt.Sprint()")
	g.writeLn("_ io.Closer")
	g.writeLn("_ net.Conn")
	g.writeLn("_ time.Time")
	g.writeLn("_ strings.Builder")
	g.writeLn("_ sync.Locker")
	g.writeLn("_ oclient.Client")
	g.writeLn("_ closer.Closer")
	g.writeLn("_ codec.Codec")
	g.writeLn("_ = packet.MaxSize")
	g.writeLn("_ oservice.Service")
	g.writeLn("_ transport.Transport")
	g.writeLn("_ validator.StructLevel")
	g.writeLn(")")
	g.writeLn("")

	// Generate the msgp shims.
	g.genTimeDurationMsgp()

	// Generate the errors.
	g.writeLn("//##############//")
	g.writeLn("//### Errors ###//")
	g.writeLn("//##############//")
	g.writeLn("")

	g.writeLn(`var ErrClosed = errors.New("closed")`)
	g.genErrors(f.Errs)

	// Generate the type definitions.
	g.writeLn("var validate = validator.New()")
	if len(f.Types) > 0 {
		g.writeLn("//#############//")
		g.writeLn("//### Types ###//")
		g.writeLn("//#############//")
		g.writeLn("")

		g.genTypes(f.Types, f.Srvc)
	}

	// Generate the enum definitions.
	if len(f.Enums) > 0 {
		g.writeLn("//#############//")
		g.writeLn("//### Enums ###//")
		g.writeLn("//#############//")
		g.writeLn("")

		g.genEnums(f.Enums)
	}

	// Generate the service definition.
	g.writeLn("//###############//")
	g.writeLn("//### Service ###//")
	g.writeLn("//###############//")
	g.writeLn("")

	g.genService(f.Srvc, f.Errs)

	return g.s.String()
}

// genTimeDurationMsgp generates a shim for the tinylib/msgp tool
// to encode/decode a time.Duration as int64.
// This allows the msgp tool to handle time.Duration.
func (g *generator) genTimeDurationMsgp() {
	g.writeLn("//### Msgp time duration shim ###//")
	g.writeLn("// See https://github.com/desertbit/orbit/issues/50")
	g.writeLn("")
	g.writeLn("//msgp:shim time.Duration as:int64 using:_encodeTimeDuration/_decodeTimeDuration")
	g.writeLn("func _encodeTimeDuration(d time.Duration) int64 {")
	g.writeLn("return int64(d)")
	g.writeLn("}")
	g.writeLn("func _decodeTimeDuration(i int64) time.Duration {")
	g.writeLn("return time.Duration(i)")
	g.writeLn("}")
	g.writeLn("")
}

// writeTimeoutParam is a helper to determine which timeout param must be written
// based on the given pointer. It automatically handles the special cases
// like no timeout or default timeout.
func (g *generator) writeTimeoutParam(timeout *time.Duration) {
	if timeout != nil {
		if *timeout == 0 {
			g.write("oservice.NoTimeout")
		} else {
			g.writef("%d*time.Nanosecond", timeout.Nanoseconds())
		}
	} else {
		g.write("oservice.DefaultTimeout")
	}
	g.write(",")
}

// writeOrbitMaxSize is a helper to determine which max size param must be written
// based on the given params. It automatically handles the special cases
// like no max size or default max size.
func (g *generator) writeOrbitMaxSizeParam(maxSize *int64, service bool) {
	imp := "oclient"
	if service {
		imp = "oservice"
	}

	if maxSize != nil {
		if *maxSize == -1 {
			g.writef("%s.NoMaxSizeLimit", imp)
		} else {
			g.write(strconv.FormatInt(*maxSize, 10))
		}
	} else {
		g.writef("%s.DefaultMaxSize", imp)
	}
	g.write(",")
}

// writeValErrCheck is a helper that writes a validate error check to the generator.
// It only does so, if the type is a struct.
func (g *generator) writeValErrCheck(dt ast.DataType, varName string) {
	// Only call validate, if it is a struct.
	if _, ok := dt.(*ast.StructType); !ok {
		return
	}

	// Validate a struct.
	g.writefLn("err = validate.Struct(%s)", varName)
	g.errIfNilFunc(func() {
		g.writefLn("err = %s(err)", valErrorCheck)
		g.writeLn("return")
	})
}

func (g *generator) errIfNil() {
	g.writeLn("if err != nil { return }")
}

func (g *generator) errIfNilFunc(f func()) {
	g.writeLn("if err != nil {")
	f()
	g.writeLn("}")
}

func (g *generator) writeLn(s string) {
	g.write(s)
	g.s.WriteString("\n")
}

func (g *generator) write(s string) {
	g.s.WriteString(s)
}

func (g *generator) writefLn(format string, a ...interface{}) {
	g.writef(format, a...)
	g.s.WriteString("\n")
}

func (g *generator) writef(format string, a ...interface{}) {
	g.s.WriteString(fmt.Sprintf(format, a...))
}
