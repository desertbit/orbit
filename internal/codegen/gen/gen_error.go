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
}

func (g *generator) genErrCheckOrbitCaller(errs []*ast.Error) {
	g.writeLn("if err != nil {")
	// Check, if a control.ErrorCode has been returned.
	if len(errs) > 0 {
		g.writeLn("var cErr *orbit.ErrorCode")
		g.writeLn("if errors.As(err, &cErr) {")
		g.writeLn("switch cErr.Code {")
		for _, e := range errs {
			g.writeLn("case ErrCode%s:", e.Name)
			g.writeLn("err = Err%s", e.Name)
		}
		g.writeLn("}")
		g.writeLn("}")
	}
	g.writeLn("return")
	g.writeLn("}")
}

func (g *generator) genErrCheckOrbitHandler(errs []*ast.Error) {
	g.writeLn("if err != nil {")
	// Check, if a api error has been returned and convert it to a control.ErrorCode.
	if len(errs) > 0 {
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
