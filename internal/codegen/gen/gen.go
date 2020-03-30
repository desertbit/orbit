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
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/parse"
	"github.com/desertbit/orbit/internal/codegen/resolve"
	"github.com/desertbit/orbit/internal/codegen/token"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v3"
)

const (
	dirPerm  = 0755
	filePerm = 0666

	cacheDir       = "orbit"
	modTimesFile   = "mod_times"
	orbitSuffix    = ".orbit"
	genOrbitSuffix = "_orbit_gen.go"
	genMsgpSuffix  = "_msgp_gen.go"

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
	modified, err := checkIfModified(orbitFile, force)
	if err != nil || !modified {
		return
	}

	// Open the file for reading.
	f, err := os.Open(orbitFile)
	if err != nil {
		return
	}
	defer f.Close()

	// Wrap the file in the token reader.
	tr := token.NewReader(bufio.NewReader(f))

	// Parse the file.
	tree, err := parse.NewParser().Parse(tr)
	if err != nil {
		return
	}

	// Resolve the whole ast.
	err = resolve.Resolve(tree)
	if err != nil {
		return
	}

	// The name of the generated file is the same as the orbit file,
	// buf with a different file ending.
	ofp := strings.TrimSuffix(orbitFile, orbitSuffix) + genOrbitSuffix

	// Generate the code into a single file.
	pkgName := filepath.Base(filepath.Dir(orbitFile))
	err = ioutil.WriteFile(ofp, []byte(generate(pkgName, tree)), filePerm)
	if err != nil {
		return
	}

	// Format the file and simplify the code, where possible.
	err = execCmd("gofmt", "-s", "-w", ofp)
	if err != nil {
		return
	}

	// Generate msgp code for it.
	mfp := strings.TrimSuffix(orbitFile, orbitSuffix) + genMsgpSuffix
	err = execCmd("msgp", "-file", ofp, "-o", mfp)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			err = errors.New("msgp required to generate MessagePack code")
		}
		return
	}
	return
}

func checkIfModified(orbitFile string, force bool) (modified bool, err error) {
	fileModTimes := make(map[string]time.Time)

	// Retrieve our cache dir.
	ucd, err := os.UserCacheDir()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("unable to retrieve cache dir, all files will be generated")
		err = nil
	} else {
		// Ensure, our directory exists.
		err = os.MkdirAll(filepath.Join(ucd, cacheDir), dirPerm)
		if err != nil {
			err = fmt.Errorf("failed to create orbit cache dir: %v", err)
			return
		}

		// Read the data from the cache file.
		var data []byte
		data, err = ioutil.ReadFile(filepath.Join(ucd, cacheDir, modTimesFile))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return
		}

		// Parse it to our struct using yaml.
		err = yaml.Unmarshal(data, &fileModTimes)
		if err != nil {
			return
		}
	}

	// Retrieve the file's info.
	fi, err := os.Lstat(orbitFile)
	if err != nil {
		return
	}

	// The file counts as modified, if its last modification timestamp
	// does not match our saved modification time, if its generated file
	// does not exist or if force is enabled.
	var exists bool
	exists, err = fileExists(strings.Replace(orbitFile, orbitSuffix, genOrbitSuffix, 1))
	if err != nil {
		return
	}
	lastModified, ok := fileModTimes[orbitFile]

	modified = !ok || !lastModified.Equal(fi.ModTime()) || force || !exists
	if modified {
		fileModTimes[orbitFile] = fi.ModTime()
	}

	// Save the updated mod times to the cache dir.
	if !force && ucd != "" {
		var data []byte
		data, err = yaml.Marshal(fileModTimes)
		if err != nil {
			return
		}

		err = ioutil.WriteFile(filepath.Join(ucd, cacheDir, modTimesFile), data, filePerm)
		if err != nil {
			return
		}
	}

	return
}

type generator struct {
	s strings.Builder
}

func generate(pkgName string, tree *ast.Tree) string {
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
		g.writeLn(imp[0] + " \"" + imp[1] + "\"")
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

	// Generate the errors.
	g.writeLn("//##############//")
	g.writeLn("//### Errors ###//")
	g.writeLn("//##############//")
	g.writeLn("")

	g.writeLn("var ErrClosed = errors.New(\"closed\")")
	g.genErrors(tree.Errs)

	// Generate the type definitions.
	g.writeLn("var validate = validator.New()")
	if len(tree.Types) > 0 {
		g.writeLn("//#############//")
		g.writeLn("//### Types ###//")
		g.writeLn("//#############//")
		g.writeLn("")

		g.genTypes(tree.Types, tree.Srvc)
	}

	// Generate the enum definitions.
	if len(tree.Enums) > 0 {
		g.writeLn("//#############//")
		g.writeLn("//### Enums ###//")
		g.writeLn("//#############//")
		g.writeLn("")

		g.genEnums(tree.Enums)
	}

	// Generate the service definition.
	g.writeLn("//###############//")
	g.writeLn("//### Service ###//")
	g.writeLn("//###############//")
	g.writeLn("")

	g.genService(tree.Srvc, tree.Errs)

	return g.s.String()
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
