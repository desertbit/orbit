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

	"github.com/desertbit/orbit/internal/codegen/ast"
)

const (
	orbitErrorCodeCheck      = "_orbitErrCodeCheck"
	errToOrbitErrorCodeCheck = "_errToOrbitErrCodeCheck"
	valErrorCheck            = "_valErrCheck"
)

func (g *generator) genErrors(errs []*ast.Error) {
	// Sort the errors in alphabetical order.
	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Name < errs[j].Name
	})

	// Write error codes.
	g.writeLn("const (")
	for _, e := range errs {
		g.writeLn("ErrCode%s = %d", e.Name, e.ID)
	}
	g.writeLn(")")

	// Write standard error variables along with the orbit Error ones.
	g.writeLn("var (")
	for _, e := range errs {
		g.writeLn("Err%s = errors.New(\"%s\")", e.Name, strExplode(e.Name))
		g.writeLn("orbitErr%s = orbit.Err(Err%s, Err%s.Error(), ErrCode%s)", e.Name, e.Name, e.Name, e.Name)
	}
	g.writeLn(")")
	g.writeLn("")

	// Generate error check helper funcs.
	g.genOrbitErrorCodeCheckFunc(errs)
	g.genErrToOrbitErrCodeFunc(errs)
	g.genValErrCheckFunc()
}

func (g *generator) genOrbitErrorCodeCheckFunc(errs []*ast.Error) {
	g.writeLn("func %s(err error) error {", orbitErrorCodeCheck)
	// Check, if a control.ErrorCode has been returned.
	g.writeLn("var cErr *orbit.ErrorCode")
	g.writeLn("if errors.As(err, &cErr) {")
	g.writeLn("switch cErr.Code {")
	for _, e := range errs {
		g.writeLn("case ErrCode%s:", e.Name)
		g.writeLn("return Err%s", e.Name)
	}
	g.writeLn("}")
	g.writeLn("}")
	g.writeLn("return err")
	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genErrToOrbitErrCodeFunc(errs []*ast.Error) {
	// Check, if one of our errors has been returned and convert it to an orbit ErrorCode.
	g.writeLn("func %s(err error) error {", errToOrbitErrorCodeCheck)
	if len(errs) > 0 {
		for i, e := range errs {
			g.writeLn("if errors.Is(err, Err%s) {", e.Name)
			g.writeLn("return orbitErr%s", e.Name)
			if i < len(errs)-1 {
				g.write("} else ")
			} else {
				g.writeLn("}")
			}
		}
	}
	g.writeLn("return err")
	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genValErrCheckFunc() {
	g.writeLn("func %s(err error) error {", valErrorCheck)
	g.writeLn("if vErrs, ok := err.(validator.ValidationErrors); ok {")
	g.writeLn("var errMsg strings.Builder")
	g.writeLn("for _, err := range vErrs {")
	g.writeLn("errMsg.WriteString(fmt.Sprintf(\"-> name: '%s', value: '%s', tag: '%s'\", err.StructNamespace(), err.Value(), err.Tag()))")
	g.writeLn("}")
	g.writeLn("return errors.New(errMsg.String())")
	g.writeLn("}")
	g.writeLn("return err")
	g.writeLn("}")
	g.writeLn("")
}
