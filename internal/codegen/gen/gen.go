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
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/desertbit/orbit/internal/codegen/lexer"
	"github.com/desertbit/orbit/internal/codegen/parser"
	"github.com/desertbit/orbit/internal/codegen/validate"
	"github.com/rs/zerolog/log"
)

const (
	dirPerm  = 0o755
	filePerm = 0o666

	orbitSuffix       = ".orbit"
	genOrbitSuffix    = "_orbit_gen.go"
	genMsgpSuffix     = "_msgp_gen.go"
	genMsgpTestSuffix = "_msgp_gen_test.go"

	recv = "v1"
)

type Config struct {
	// The directory into which QML types should be generated.
	// Type generation for QML is disabled, if this dir is empty.
	QMLDir string
	// A list of type names from the orbit file that should be ignored when generating QML types.
	QMLSkipTypes []string

	// If force is true, the cache is ignored.
	Force bool
}

// Generate processes the given orbit file and generates the go code for it into the same directory.
func Generate(orbitFile string, cfg Config) (err error) {
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
	modified, err := compareWithGenCache(orbitFile, cfg.Force)
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
	input, err := os.ReadFile(orbitFile)
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

	// Generate the Go code into a single file.
	pkgName := filepath.Base(filepath.Dir(orbitFile))
	err = os.WriteFile(ofp, []byte(goGenerate(pkgName, f)), filePerm)
	if err != nil {
		return
	}

	// Format the file and simplify the code, where possible.
	err = execCmd("gofmt", "-s", "-w", ofp)
	if err != nil {
		return
	}

	// If requested, generate the QML type definitions.
	if cfg.QMLDir != "" {
		err = os.MkdirAll(cfg.QMLDir, 0o755)
		if err != nil {
			return fmt.Errorf("mkdir all (path=%s): %v", cfg.QMLDir, err)
		}

		err = qmlGenerate(cfg.QMLDir, cfg.QMLSkipTypes, f)
		if err != nil {
			return fmt.Errorf("qml generate: %v", err)
		}
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
			if !errors.Is(err, fs.ErrNotExist) {
				return
			}
			err = nil
		}
		err = os.Remove(filePathNoSuffix + genMsgpTestSuffix)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return
			}
			err = nil
		}
	}
	return
}

const indentStr = "    " // 4 Spaces.

type generator struct {
	s strings.Builder

	// Number of indentation to apply to current write commands.
	indentLevel int
	// True, if the current line has been indented.
	currentLineHasIndent bool
}

func (g *generator) writeLn(s string) {
	g.applyIndent()
	g.write(s)
	g.s.WriteString("\n")
	g.resetIndent()
}

func (g *generator) write(s string) {
	g.applyIndent()
	g.s.WriteString(s)
}

func (g *generator) writefLn(format string, a ...any) {
	g.applyIndent()
	g.writef(format, a...)
	g.s.WriteString("\n")
	g.resetIndent()
}

func (g *generator) writef(format string, a ...any) {
	g.applyIndent()
	g.s.WriteString(fmt.Sprintf(format, a...))
}

// Do not use on its own.
func (g *generator) applyIndent() {
	if !g.currentLineHasIndent {
		g.s.WriteString(strings.Repeat(indentStr, g.indentLevel))
		g.currentLineHasIndent = true
	}
}

func (g *generator) resetIndent() {
	g.currentLineHasIndent = false
}

func (g *generator) indent(f func()) {
	g.indentLevel++
	g.currentLineHasIndent = false
	f()
	g.indentLevel--
}
