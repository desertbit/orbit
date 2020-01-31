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

func (g *generator) genServiceCallCallerSignature(c *ast.Call) {
	g.write("%s(ctx context.Context", c.Name)
	if c.Args != nil {
		g.write(", args %s", c.Args.Decl())
	}
	g.write(") (")
	if c.Ret != nil {
		g.write("ret %s, ", c.Ret.Decl())
	}
	g.write("err error)")
}

func (g *generator) genServiceCallCaller(c *ast.Call, srvcName, structName string, errs []*ast.Error) {
	// Method declaration.
	g.write("func (%s *%s) ", recv, structName)
	g.genServiceCallCallerSignature(c)
	g.writeLn(" {")

	// Method body.
	// Set the timeout, if needed.
	if c.Timeout != nil && *c.Timeout > 0 {
		g.writeLn("ctx, cancel := context.WithTimeout(ctx, %d*time.Nanosecond)", c.Timeout.Nanoseconds())
		g.writeLn("defer cancel()")
	} else {
		g.writeLn("ct := %s.s.CallTimeout()", recv)
		g.writeLn("if ct > 0 {")
		g.writeLn("var cancel context.CancelFunc")
		g.writeLn("ctx, cancel = context.WithTimeout(ctx, ct)")
		g.writeLn("defer cancel()")
		g.writeLn("}")
	}

	// Make the call.
	if c.Ret != nil {
		g.write("retData, err := ")
	} else {
		g.write("_, err = ")
	}
	g.write("%s.s.Call", recv)
	if c.Async {
		g.write("Async")
	}
	g.write("(ctx, Service%s, %s, ", srvcName, srvcName+c.Name)
	if c.Args != nil {
		g.writeLn("args)")
	} else {
		g.writeLn("nil)")
	}

	// Check error and parse control.ErrorCodes.
	g.errIfNilFunc(func() {
		g.writeLn("err = %s(err)", orbitErrorCodeCheck)
		g.writeLn("return")
	})

	// If return arguments are expected, decode them.
	if c.Ret != nil {
		// Parse.
		g.writeLn("err = retData.Decode(&ret)")
		g.errIfNil()

		// Validate.
		if c.RetValTag != "" {
			// Validate a single value.
			g.writeLn("err = validate.Var(ret, \"%s\")", c.RetValTag)
		} else {
			// Validate a struct.
			g.writeLn("err = validate.Struct(ret)")
		}
		g.errIfNilFunc(func() {
			g.writeLn("err = %s(err)", valErrorCheck)
			g.writeLn("return")
		})
	}

	// Return.
	g.writeLn("return")

	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceCallHandlerSignature(c *ast.Call) {
	g.write("%s(ctx context.Context, s *orbit.Session", c.Name)
	if c.Args != nil {
		g.write(", args %s", c.Args.Decl())
	}
	g.write(") (")
	if c.Ret != nil {
		g.write("ret %s, ", c.Ret.Decl())
	}
	g.write("err error)")
}

func (g *generator) genServiceCallHandler(c *ast.Call, structName string, errs []*ast.Error) {
	// Method declaration.
	g.writeLn(
		"func (%s *%s) %s(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {",
		recv, structName, c.NamePrv(),
	)

	// Method body.
	// Parse and validate the args.
	handlerArgs := "ctx, s"
	if c.Args != nil {
		handlerArgs += ", args"

		// Parse.
		g.writeLn("var args %s", c.Args.Decl())
		g.writeLn("err = ad.Decode(&args)")
		g.errIfNil()

		// Validate.
		if c.ArgsValTag != "" {
			// Validate a single value.
			g.writeLn("err = validate.Var(args, \"%s\")", c.ArgsValTag)
		} else {
			// Validate a struct.
			g.writeLn("err = validate.Struct(args)")
		}
		g.errIfNilFunc(func() {
			g.writeLn("err = %s(err)", valErrorCheck)
			g.writeLn("return")
		})
	}

	// Call the handler.
	if c.Ret != nil {
		g.writeLn("ret, err := %s.h.%s(%s)", recv, c.Name, handlerArgs)
	} else {
		g.writeLn("err = %s.h.%s(%s)", recv, c.Name, handlerArgs)
	}

	// Check error and convert to orbit errors.
	g.errIfNilFunc(func() {
		g.writeLn("err = %s(err)", errToOrbitErrorCodeCheck)
		g.writeLn("return")
	})

	// Assign return value.
	if c.Ret != nil {
		g.writeLn("r = ret")
	}

	// Return.
	g.writeLn("return")

	g.writeLn("}")
	g.writeLn("")
}
