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
	valErrorCheck = "_valErrCheck"
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
			g.writefLn("ErrCode%s = %d", e.Ident(), e.ID)
		}
		g.writeLn(")")

		// Write standard error variables.
		g.writeLn("var (")
		for _, e := range errs {
			g.writefLn("Err%s = errors.New(\"%s\")", e.Ident(), strExplode(e.Ident()))
		}
		g.writeLn(")")
		g.writeLn("")
	}

	// Generate error check helper funcs.
	g.genValErrCheckFunc()
}

func (g *generator) genClientErrorInlineCheck(errs []*ast.Error) {
	g.writeLn("var cErr oclient.Error")
	g.writeLn("if errors.As(err, &cErr) {")
	g.writeLn("switch cErr.Code() {")
	for _, e := range errs {
		g.writefLn("case ErrCode%s:", e.Ident())
		g.writefLn("err = Err%s", e.Ident())
	}
	g.writeLn("}")
	g.writeLn("}")
	g.writeLn("return")
}

func (g *generator) genServiceErrorInlineCheck(errs []*ast.Error) {
	for i, e := range errs {
		g.writefLn("if errors.Is(err, Err%s) {", e.Ident())
		g.writefLn("err = oservice.NewError(err, Err%s.Error(), ErrCode%s)", e.Ident(), e.Ident())
		if i < len(errs)-1 {
			g.write("} else ")
		} else {
			g.writeLn("}")
		}
	}
	g.writeLn("return")
}

func (g *generator) genValErrCheckFunc() {
	g.writefLn("func %s(err error) error {", valErrorCheck)
	g.writeLn("if vErrs, ok := err.(validator.ValidationErrors); ok {")
	g.writeLn("var errMsg strings.Builder")
	g.writeLn("for _, err := range vErrs {")
	g.writeLn("errMsg.WriteString(fmt.Sprintf(\"[name: '%s', value: '%s', tag: '%s']\", err.StructNamespace(), err.Value(), err.Tag()))")
	g.writeLn("}")
	g.writeLn("return errors.New(errMsg.String())")
	g.writeLn("}")
	g.writeLn("return err")
	g.writeLn("}")
	g.writeLn("")
}
