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
	"github.com/desertbit/orbit/internal/codegen/ast"
)

func (g *generator) genServiceCallCallerSignature(c *ast.Call, srvcName string) {
	write("%s(ctx context.Context", srvcName+c.Name)
	if c.Args != nil {
		write(", args %s", c.Args.String())
	}
	write(") (")
	if c.Ret != nil {
		write("ret %s, ", c.Ret.String())
	}
	write("err error)")
}

func (g *generator) genServiceCallHandlerSignature(c *ast.Call, srvcName string) {
	write("%s(ctx context.Context, s *orbit.Session", srvcName+c.Name)
	if c.Args != nil {
		write(", args %s", c.Args.String())
	}
	write(") (")
	if c.Ret != nil {
		write("ret %s, ", c.Ret.String())
	}
	write("err error)")
}

func (g *generator) genServiceCallClient(c *ast.Call, srvcName, structName string, errs []*ast.Error) {
	// Method declaration.
	write("func (%s *%s) ", recv, structName)
	g.genServiceCallCallerSignature(c, srvcName)
	writeLn(" {")

	// Method body.
	// First, make the call.
	if c.Ret != nil {
		write("retData, err := ")
	} else {
		write("_, err = ")
	}
	write("%s.s.Call", recv)
	if c.Async {
		write("Async")
	}
	write("(ctx, Service%s, %s, ", srvcName, srvcName+c.Name)
	if c.Args != nil {
		writeLn("args)")
	} else {
		writeLn("nil)")
	}

	// Check error and parse control.ErrorCodes.
	genErrCheckOrbitCaller(errs)

	// If return arguments are expected, decode them.
	if c.Ret != nil {
		writeLn("ret = &%s{}", c.Ret.Name)
		writeLn("err = retData.Decode(ret)")
		errIfNil()
	}

	// Return.
	writeLn("return")

	writeLn("}")
	writeLn("")
}

func (g *generator) genServiceCallServer(c *ast.Call, srvcName, srvcNamePrv, structName string, errs []*ast.Error) {
	// Method declaration.
	writeLn(
		"func (%s *%s) %s(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {",
		recv, structName, srvcNamePrv+c.Name,
	)

	// Method body.
	// Parse the args.
	handlerArgs := "ctx, s"
	if c.Args != nil {
		handlerArgs += ", args"
		writeLn("args := &%s{}", c.Args.Name)
		writeLn("err = ad.Decode(args)")
		errIfNil()
	}

	// Call the handler.
	if c.Ret != nil {
		writeLn("ret, err := %s.h.%s(%s)", recv, srvcName+c.Name, handlerArgs)
	} else {
		writeLn("err = %s.h.%s(%s)", recv, srvcName+c.Name, handlerArgs)
	}

	// Check error and convert to orbit errors.
	genErrCheckOrbitHandler(errs)

	// Assign return value.
	if c.Ret != nil {
		writeLn("r = ret")
	}

	// Return.
	writeLn("return")

	writeLn("}")
	writeLn("")
}
