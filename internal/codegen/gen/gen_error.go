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
	clientErrorCheck  = "_clientErrorCheck"
	serviceErrorCheck = "_serviceErrorCheck"
	valErrorCheck     = "_valErrCheck"
)

func (g *generator) genErrors(errs []*ast.Error) {
	// Sort the errors in alphabetical order.
	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Name < errs[j].Name
	})

	// Write error codes.
	if len(errs) > 0 {
		g.writeLn("const (")
		for _, e := range errs {
			g.writeLn("ErrCode%s = %d", e.Name, e.ID)
		}
		g.writeLn(")")

		// Write standard error variables along with the service Error ones.
		g.writeLn("var (")
		for _, e := range errs {
			g.writeLn("Err%s = errors.New(\"%s\")", e.Name, strExplode(e.Name))
			g.writeLn("serviceErr%s = oservice.Err(Err%s, Err%s.Error(), ErrCode%s)", e.Name, e.Name, e.Name, e.Name)
		}
		g.writeLn(")")
		g.writeLn("")
	}

	// Generate error check helper funcs.
	g.genClientErrorCheckFunc(errs)
	g.genServiceErrorCheckFunc(errs)
	g.genValErrCheckFunc()
}

func (g *generator) genClientErrorCheckFunc(errs []*ast.Error) {
	g.writeLn("func %s(err error) error {", clientErrorCheck)
	if len(errs) == 0 {
		g.writeLn("return err")
		g.writeLn("}")
		return
	}

	// Check, if a client.Error has been returned.
	g.writeLn("var cErr oclient.Error")
	g.writeLn("if errors.As(err, &cErr) {")
	g.writeLn("switch cErr.Code() {")
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

func (g *generator) genServiceErrorCheckFunc(errs []*ast.Error) {
	// Check, if one of our errors has been returned and convert it to a service Error.
	g.writeLn("func %s(err error) error {", serviceErrorCheck)
	if len(errs) == 0 {
		g.writeLn("return err")
		g.writeLn("}")
		return
	}

	for i, e := range errs {
		g.writeLn("if errors.Is(err, Err%s) {", e.Name)
		g.writeLn("return serviceErr%s", e.Name)
		if i < len(errs)-1 {
			g.write("} else ")
		} else {
			g.writeLn("}")
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
	g.writeLn("errMsg.WriteString(\"-> name: '\"+err.StructNamespace()+\"', value: '\"+err.Value()+\"', tag: '\"+err.Tag()+\"'\"")
	g.writeLn("}")
	g.writeLn("return errors.New(errMsg.String())")
	g.writeLn("}")
	g.writeLn("return err")
	g.writeLn("}")
	g.writeLn("")
}
