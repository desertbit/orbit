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

/*
TODO: clean up generator struct, it was originally designed for multiple and single files.
*/

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

func Generate(dir string, force bool) (err error) {
	// Ensure, the directory's path is absolute!
	dir, err = filepath.Abs(dir)
	if err != nil {
		return
	}

	// Find all file paths of orbit files an that dir.
	orbitFilePaths, err := findModifiedOrbitFiles(dir, force)
	if err != nil || len(orbitFilePaths) == 0 {
		return
	}

	var (
		tr   = token.NewReader(nil)
		p    = parse.NewParser()
		tree = &ast.Tree{}
	)

	// Parse all found files.
	for _, fp := range orbitFilePaths {
		err = func() (err error) {
			// Open file for reading.
			var f *os.File
			f, err = os.Open(fp)
			if err != nil {
				return
			}
			defer f.Close()

			// Wrap the file in the token reader.
			tr.Reset(bufio.NewReader(f))

			// Parse the file.
			var tree2 *ast.Tree
			tree2, err = p.Parse(tr)
			if err != nil {
				return
			}

			// Combine the tree.
			tree.Srvcs = append(tree.Srvcs, tree2.Srvcs...)
			tree.Types = append(tree.Types, tree2.Types...)
			tree.Errs = append(tree.Errs, tree2.Errs...)
			tree.Enums = append(tree.Enums, tree2.Enums...)
			return
		}()
		if err != nil {
			return
		}
	}

	// Resolve the whole ast.
	err = resolve.Resolve(tree)
	if err != nil {
		return
	}

	// Generate the code into a single file.
	// The name of the generated file is the package name.
	pkgName := filepath.Base(dir)
	ofp := filepath.Join(dir, pkgName+genOrbitSuffix)
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
	mfp := strings.TrimSuffix(ofp, genOrbitSuffix) + genMsgpSuffix
	err = execCmd("msgp", "-file", ofp, "-o", mfp)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			err = errors.New("msgp required to generate MessagePack code")
		}
		return
	}

	return
}

func findModifiedOrbitFiles(dir string, force bool) (filePaths []string, err error) {
	fileModTimes := make(map[string]time.Time)

	// Retrieve our cache dir.
	ucd, err := os.UserCacheDir()
	if err != nil {
		log.Warn().Err(err).Msg("unable to retrieve cache dir, all files will be generated")
		err = nil
	} else {
		// Ensure, our directory exists.
		err = os.MkdirAll(filepath.Join(ucd, cacheDir), dirPerm)
		if err != nil {
			err = fmt.Errorf("failed to create orbit config dir: %v", err)
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

	// Gather the info about all files in the dir.
	info, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	// Find all orbit files and collect their file paths.
	// At the same time, if we find at least one file that matches
	// our criteria of being modified, we will return the file paths,
	// so that the code is regenerated.
	var foundOneMod bool
	for _, fi := range info {
		fileName := fi.Name()

		// Check, if the file is an orbit file.
		if !fi.Mode().IsRegular() || !strings.HasSuffix(fileName, orbitSuffix) {
			continue
		}

		// Add to the orbit file paths.
		fp := filepath.Join(dir, fileName)
		filePaths = append(filePaths, fp)

		// A file counts as modified, if its last modification timestamp
		// does not match our saved modification time, if its generated file
		// does not exist, or if force is enabled.

		var exists bool
		exists, err = fileExists(filepath.Join(dir, filepath.Base(dir)+genOrbitSuffix))
		if err != nil {
			return
		}

		// Check, if the file will be regenerated. If so, save also its
		// modification time in our map.
		lastModified, ok := fileModTimes[fp]
		if !ok || !lastModified.Equal(fi.ModTime()) || force || !exists {
			fileModTimes[fp] = fi.ModTime()
			foundOneMod = true
		}
	}

	// Save the updated mod times to the cache dir, if force is not enabled.
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

	// If not a single file has been modified, do not return any paths.
	if !foundOneMod {
		filePaths = make([]string, 0)
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
	g.writeLn("package %s", pkgName)
	g.writeLn("")

	// Write the imports.
	imports := [][2]string{
		{"context", "context"},
		{"errors", "errors"},
		{"fmt", "fmt"},
		{"net", "net"},
		{"time", "time"},
		{"sync", "sync"},
		{"io", "io"},
		{"strings", "strings"},
		{"validator", "github.com/go-playground/validator/v10"},
		{"orbit", "github.com/desertbit/orbit/pkg/orbit"},
		{"codec", "github.com/desertbit/orbit/pkg/codec"},
		{"packet", "github.com/desertbit/orbit/pkg/packet"},
		{"closer", "github.com/desertbit/closer/v3"},
	}

	g.writeLn("import (")
	for _, imp := range imports {
		g.writeLn(imp[0] + " \"" + imp[1] + "\"")
	}
	g.writeLn(")")
	g.writeLn("")

	// Generate a var block that imports one struct from every package to ensure usage.
	g.writeLn("var (")
	g.writeLn("_ context.Context")
	g.writeLn("_ = errors.New(\"\")")
	g.writeLn("_ net.Conn")
	g.writeLn("_ time.Time")
	g.writeLn("_ sync.Locker")
	g.writeLn("_ orbit.Conn")
	g.writeLn("_ = fmt.Sprint()")
	g.writeLn("_ = strings.Builder{}")
	g.writeLn("_ = io.EOF")
	g.writeLn("_ validator.StructLevel")
	g.writeLn("_ codec.Codec")
	g.writeLn("_ = packet.MaxSize")
	g.writeLn("_ closer.Closer")
	g.writeLn(")")
	g.writeLn("")

	// Generate the errors.
	g.writeLn("//##############//")
	g.writeLn("//### Errors ###//")
	g.writeLn("//##############//")
	g.writeLn("")

	g.writeLn("var ErrClosed = errors.New(\"closed\")")
	if len(tree.Errs) > 0 {
		g.genErrors(tree.Errs)
	}

	// Generate the type definitions.
	g.writeLn("var validate = validator.New()")
	if len(tree.Types) > 0 {
		g.writeLn("//#############//")
		g.writeLn("//### Types ###//")
		g.writeLn("//#############//")
		g.writeLn("")

		g.genTypes(tree.Types, tree.Srvcs)
	}

	// Generate the enum definitions.
	if len(tree.Enums) > 0 {
		g.writeLn("//#############//")
		g.writeLn("//### Enums ###//")
		g.writeLn("//#############//")
		g.writeLn("")

		g.genEnums(tree.Enums)
	}

	// Generate the service definitions.
	if len(tree.Srvcs) > 0 {
		g.writeLn("//################//")
		g.writeLn("//### Services ###//")
		g.writeLn("//################//")
		g.writeLn("")

		g.genServices(tree.Srvcs, tree.Errs)
	}

	return g.s.String()
}

func (g *generator) errIfNil() {
	g.writeLn("if err != nil { return }")
}

func (g *generator) errIfNilFunc(f func()) {
	g.writeLn("if err != nil {")
	f()
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
