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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/desertbit/orbit/internal/parse"
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

func Generate(dir string, streamChanSize uint, splitOutput, force bool) (err error) {
	orbitFilePaths, err := findModifiedOrbitFiles(dir, splitOutput, force)
	if err != nil || len(orbitFilePaths) == 0 {
		return
	}

	var (
		errs         []*parse.Error
		services     []*parse.Service
		types        []*parse.Type
		genFilePaths []string

		pkgName = filepath.Base(dir)
		g       = &generator{}
	)

	if splitOutput {
		var data []byte

		for _, fp := range orbitFilePaths {
			// Read the file data.
			data, err = ioutil.ReadFile(fp)
			if err != nil {
				return
			}

			// Parse the data.
			services, types, errs, err = parse.Parse(data)
			if err != nil {
				err = fmt.Errorf("parsing failed\n-> %v\n-> %s", err, fp)
				return
			}

			// The output file has the same name as its orbit file.
			// Add its path to the generated files.
			genFilePath := filepath.Join(dir, strings.TrimSuffix(filepath.Base(fp), orbitSuffix)+genOrbitSuffix)
			genFilePaths = append(genFilePaths, genFilePath)

			err = ioutil.WriteFile(
				genFilePath,
				[]byte(g.genHeader(pkgName)+g.genBody(errs, types, services, streamChanSize)),
				filePerm,
			)
			if err != nil {
				return
			}
		}
	} else {
		var input []byte

		// First, read in all files.
		for _, fp := range orbitFilePaths {
			// Read the file data.
			var data []byte
			data, err = ioutil.ReadFile(fp)
			if err != nil {
				return
			}

			input = append(input, data...)
		}

		// The single output file has the name of the directory as prefix.
		// Add its path to the generated files.
		genFilePath := filepath.Join(dir, filepath.Base(dir)+genOrbitSuffix)
		genFilePaths = append(genFilePaths, genFilePath)

		// Parse all files in one go.
		services, types, errs, err = parse.Parse(input)
		if err != nil {
			return
		}

		// Write the output to the file.
		err = ioutil.WriteFile(
			genFilePath,
			[]byte(g.genHeader(pkgName)+g.genBody(errs, types, services, streamChanSize)),
			filePerm,
		)
		if err != nil {
			return
		}
	}

	// Now, format the files and generate the msgp code for them.
	for _, gfp := range genFilePaths {
		// Format the file.
		err = execCmd("gofmt", "-s", "-w", gfp)
		if err != nil {
			return
		}

		// Generate msgp code for it.
		err = execCmd("msgp", "-file", gfp, "-o", strings.TrimSuffix(gfp, genOrbitSuffix)+genMsgpSuffix)
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				err = errors.New("msgp required to generate MessagePack code")
			}
			return
		}
	}

	return
}

func findModifiedOrbitFiles(dir string, splitOutput, force bool) (modifiedFilePaths []string, err error) {
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

	// Search the orbit files in the dir.
	info, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	// Save all file paths to orbit files.
	allFilePaths := make([]string, 0)

	for _, fi := range info {
		fileName := fi.Name()

		// Check, if the file has our orbit suffix.
		if !fi.Mode().IsRegular() || !strings.HasSuffix(fileName, orbitSuffix) {
			continue
		}

		// Add to the orbit file paths.
		fp := filepath.Join(dir, fileName)
		allFilePaths = append(allFilePaths, fp)

		// In order to add the file to the modified files, it must have been
		// modified since the last generation, or force must be enabled.
		// Additionally, the generated file of it must exist.

		var (
			exists bool
			genFp  string
		)
		if splitOutput {
			genFp = strings.TrimSuffix(fp, orbitSuffix) + genOrbitSuffix
		} else {
			genFp = filepath.Join(dir, filepath.Base(dir)+genOrbitSuffix)
		}
		exists, err = fileExists(genFp)
		if err != nil {
			return
		}

		lastModified, ok := fileModTimes[fp]
		if !ok || !lastModified.Equal(fi.ModTime()) || force || !exists {
			modifiedFilePaths = append(modifiedFilePaths, fp)
			fileModTimes[fp] = fi.ModTime()
		}
	}

	// Save the updated mod times to the config dir, if force not enabled.
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

	// In case only a single output file should be generated, we must generate all orbit files
	// if at least one has been modified, or force is enabled.
	if !splitOutput && (len(modifiedFilePaths) > 0 || force) {
		modifiedFilePaths = allFilePaths
	}

	return
}

type generator struct {
	s strings.Builder
}

func (g *generator) genHeader(pkgName string) (header string) {
	// Write the preamble.
	g.writeLn("/* code generated by orbit */")
	g.writeLn("package %s", pkgName)
	g.writeLn("")

	// Write the imports.
	imports := []string{
		"context",
		"errors",
		"net",
		"time",
		"sync",
		"github.com/desertbit/orbit/pkg/orbit",
		"github.com/desertbit/orbit/pkg/packet",
		"github.com/desertbit/closer/v3",
	}

	g.writeLn("import (")
	for _, imp := range imports {
		g.writeLn("\"" + imp + "\"")
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
	g.writeLn("_ = packet.MaxSize")
	g.writeLn("_ closer.Closer")
	g.writeLn(")")
	g.writeLn("")

	// Return the header and reset the builder.
	header = g.s.String()
	g.s.Reset()
	return
}

func (g *generator) genBody(errs []*parse.Error, types []*parse.Type, services []*parse.Service, streamChanSize uint) (body string) {
	// Generate the errors.
	g.writeLn("//#####################//")
	g.writeLn("//### Global Errors ###//")
	g.writeLn("//#####################//")
	g.writeLn("")

	if len(errs) > 0 {
		g.genErrors(errs)
	}

	// Generate the type definitions.
	g.writeLn("//####################//")
	g.writeLn("//### Global Types ###//")
	g.writeLn("//####################//")
	g.writeLn("")

	if len(types) > 0 {
		g.genTypes(types, services, streamChanSize)
	}

	// Generate the service definitions.
	g.writeLn("//################//")
	g.writeLn("//### Services ###//")
	g.writeLn("//################//")
	g.writeLn("")

	if len(services) > 0 {
		g.genServices(services, errs, streamChanSize)
	}

	body = g.s.String()
	g.s.Reset()
	return
}

func (g *generator) errIfNil() {
	g.writeLn("if err != nil { return }")
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
