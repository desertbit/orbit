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

func (g *generator) genServiceClientCallSignature(c *ast.Call) {
	g.write("%s(ctx context.Context", c.Name)
	if c.Arg != nil {
		g.write(", arg %s", c.Arg.Decl())
	}
	g.write(") (")
	if c.Ret != nil {
		g.write("ret %s, ", c.Ret.Decl())
	}
	g.write("err error)")
}

func (g *generator) genServiceClientCall(c *ast.Call, errs []*ast.Error) {
	// Method declaration.
	g.write("func (%s *client) ", recv)
	g.genServiceClientCallSignature(c)
	g.writeLn(" {")

	// Method body.
	// Set the timeout, if needed.
	if c.Timeout != nil && *c.Timeout > 0 {
		g.writeLn("ctx, cancel := context.WithTimeout(ctx, %d*time.Nanosecond)", c.Timeout.Nanoseconds())
		g.writeLn("defer cancel()")
	} else {
		g.writeLn("if %s.callTimeout > 0 {", recv)
		g.writeLn("var cancel context.CancelFunc")
		g.writeLn("ctx, cancel = context.WithTimeout(ctx, %s.callTimeout)", recv)
		g.writeLn("defer cancel()")
		g.writeLn("}")
	}

	g.write("err = %s.", recv)
	if c.Async {
		g.write("Async")
	}
	g.write("Call")

	g.write("(ctx, %s, ", c.Name)
	if c.Arg != nil {
		g.write("arg,")
	} else {
		g.write("nil,")
	}

	if c.Ret != nil {
		g.write("&ret,")
	} else {
		g.write("nil,")
	}
	g.writeLn(")")

	// Check error and parse control.ErrorCodes.
	g.errIfNilFunc(func() {
		g.writeLn("err = %s(err)", clientErrorCheck)
		g.writeLn("return")
	})

	// If return arguments were expected, validate them.
	if c.Ret != nil {
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

func (g *generator) genServiceHandlerCallSignature(c *ast.Call) {
	g.write("%s(ctx oservice.Context", c.Name)
	if c.Arg != nil {
		g.write(", arg %s", c.Arg.Decl())
	}
	g.write(") (")
	if c.Ret != nil {
		g.write("ret %s, ", c.Ret.Decl())
	}
	g.write("err error)")
}

func (g *generator) genServiceHandlerCall(c *ast.Call, errs []*ast.Error) {
	// Method declaration.
	g.writeLn(
		"func (%s *service) %s(ctx oservice.Context, argData []byte) (retData interface{}, err error) {",
		recv, c.NamePrv(),
	)

	// Method body.
	// Parse and validate the args.
	handlerArgs := "ctx,"
	if c.Arg != nil {
		handlerArgs += "arg,"

		// Parse.
		g.writeLn("var arg %s", c.Arg.Decl())
		g.writeLn("err = %s.codec.Decode(argData, &arg)", recv)
		g.errIfNil()

		// Validate.
		if c.ArgValTag != "" {
			// Validate a single value.
			g.writeLn("err = validate.Var(arg, \"%s\")", c.ArgValTag)
		} else {
			// Validate a struct.
			g.writeLn("err = validate.Struct(arg)")
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
		g.writeLn("err = %s(err)", serviceErrorCheck)
		g.writeLn("return")
	})

	if c.Ret != nil {
		// Check for nil return value.
		g.writeLn("if ret == nil {")
		g.writeLn("err = errors.New(\"return value is a nil pointer\")")
		g.writeLn("return")
		g.writeLn("}")

		// Assign return value.
		if c.Ret != nil {
			g.writeLn("retData = ret")
		}
	}

	// Return.
	g.writeLn("return")

	g.writeLn("}")
	g.writeLn("")
}
