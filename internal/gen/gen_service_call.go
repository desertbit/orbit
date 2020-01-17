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
	"github.com/desertbit/orbit/internal/parse"
)

func (g *generator) genServiceCallCallerSignature(c *parse.Call) {
	g.write("%s(ctx context.Context", c.Name)
	if c.Args != nil {
		g.write(", args %s", c.Args.String())
	}
	g.write(") (")
	if c.Ret != nil {
		g.write("ret %s, ", c.Ret.String())
	}
	g.write("err error)")
}

func (g *generator) genServiceCallHandlerSignature(c *parse.Call) {
	g.write("%s(ctx context.Context, s *orbit.Session", c.Name)
	if c.Args != nil {
		g.write(", args %s", c.Args.String())
	}
	g.write(") (")
	if c.Ret != nil {
		g.write("ret %s, ", c.Ret.String())
	}
	g.write("err error)")
}

func (g *generator) genServiceCallClient(c *parse.Call, structName, srvcName string, errs []*parse.Error) {
	// Method declaration.
	g.write("func (%s *%s) ", recv, structName)
	g.genServiceCallCallerSignature(c)
	g.writeLn(" {")

	// Method body.
	// First, make the call.
	if c.Ret != nil {
		g.write("retData, err := ")
	} else {
		g.write("_, err = ")
	}
	g.write("%s.s.Call", recv)
	if c.Async {
		g.write("Async")
	}
	g.write("(ctx, %s, %s, ", srvcName, srvcName+c.Name)
	if c.Args != nil {
		g.writeLn("args)")
	} else {
		g.writeLn("nil)")
	}

	// Check error and parse control.ErrorCodes.
	g.genErrCheckOrbitCaller(errs)

	// If return arguments are expected, decode them.
	if c.Ret != nil {
		g.writeLn("ret = &%s{}", c.Ret.Name)
		g.writeLn("err = retData.Decode(ret)")
		g.errIfNil()
	}

	// Return.
	g.writeLn("return")

	g.writeLn("}")
	g.writeLn("")
}

func (g *generator) genServiceCallServer(c *parse.Call, structName string, errs []*parse.Error) {
	// Method declaration.
	g.writeLn(
		"func (%s *%s) %s(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {",
		recv, structName, c.NamePrv(),
	)

	// Method body.
	// Parse the args.
	handlerArgs := "ctx, s"
	if c.Args != nil {
		handlerArgs += ", args"
		g.writeLn("args := &%s{}", c.Args.Name)
		g.writeLn("err = ad.Decode(args)")
		g.errIfNil()
	}

	// Call the handler.
	if c.Ret != nil {
		g.writeLn("ret, err := %s.h.%s(%s)", recv, c.Name, handlerArgs)
	} else {
		g.writeLn("err = %s.h.%s(%s)", recv, c.Name, handlerArgs)
	}

	// Check error and convert to orbit errors.
	g.genErrCheckOrbitHandler(errs)

	// Assign return value.
	if c.Ret != nil {
		g.writeLn("r = ret")
	}

	// Return.
	g.writeLn("return")

	g.writeLn("}")
	g.writeLn("")
}
