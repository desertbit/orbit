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

package resolve

import (
	"github.com/desertbit/orbit/internal/codegen/ast"
)

func resolveService(srvc *ast.Service, types []*ast.Type, enums []*ast.Enum) (err error) {
	for j, c := range srvc.Calls {
		for k := j + 1; k < len(srvc.Calls); k++ {
			// Check for duplicate names.
			if c.Name == srvc.Calls[k].Name {
				return ast.NewErr(c.Line, "call '%s' declared twice", c.Name)
			}
		}

		// Resolve the call.
		err = resolveCall(c, types, enums)
		if err != nil {
			return
		}
	}

	for j, s := range srvc.Streams {
		for k := j + 1; k < len(srvc.Streams); k++ {
			// Check for duplicate names.
			if s.Name == srvc.Streams[k].Name {
				return ast.NewErr(s.Line, "stream '%s' declared twice", s.Name)
			}
		}

		// Resolve the stream.
		err = resolveStream(s, types, enums)
		if err != nil {
			return
		}
	}

	return
}

func resolveCall(c *ast.Call, types []*ast.Type, enums []*ast.Enum) (err error) {
	// Resolve all AnyTypes.
	c.Arg, err = resolveAnyType(c.Arg, types, enums)
	if err != nil {
		return
	}
	c.Ret, err = resolveAnyType(c.Ret, types, enums)
	if err != nil {
		return
	}

	// Check, that no StructType has a validation tag.
	if st, ok := c.Arg.(*ast.StructType); ok && c.ArgValTag != "" {
		return ast.NewErr(st.Line, "validation tags not allowed on type references")
	}
	if st, ok := c.Ret.(*ast.StructType); ok && c.RetValTag != "" {
		return ast.NewErr(st.Line, "validation tags not allowed on type references")
	}

	// Check async calls and their special properties.
	if c.Async {
		// No negative timeouts.
		if c.Timeout != nil && *c.Timeout < 0 {
			return ast.NewErr(c.Line, "negative timeouts are not allowed")
		}

		// Check, that max sizes are only set, if the respective call data is defined.
		if c.Arg == nil && c.MaxArgSize != nil {
			return ast.NewErr(c.Line, "max arg size given, but arg not defined")
		}
		if c.Ret == nil && c.MaxRetSize != nil {
			return ast.NewErr(c.Line, "max ret size given, but ret not defined")
		}
	} else {
		// For standard calls, no custom max sizes are allowed.
		if c.MaxArgSize != nil || c.MaxRetSize != nil {
			return ast.NewErr(c.Line, "only async calls can have custom max arg/ret sizes")
		}
	}

	return
}

func resolveStream(s *ast.Stream, types []*ast.Type, enums []*ast.Enum) (err error) {
	// Resolve all AnyTypes.
	s.Arg, err = resolveAnyType(s.Arg, types, enums)
	if err != nil {
		return
	}
	s.Ret, err = resolveAnyType(s.Ret, types, enums)
	if err != nil {
		return
	}

	// Check, that no StructType has a validation tag.
	if st, ok := s.Arg.(*ast.StructType); ok && s.ArgValTag != "" {
		return ast.NewErr(st.Line, "validation tags not allowed on type references")
	}
	if st, ok := s.Ret.(*ast.StructType); ok && s.RetValTag != "" {
		return ast.NewErr(st.Line, "validation tags not allowed on type references")
	}

	// Check, that max sizes are only set, if the respective stream data is defined.
	if s.Arg == nil && s.MaxArgSize != nil {
		return ast.NewErr(s.Line, "max arg size given, but arg not defined")
	}
	if s.Ret == nil && s.MaxRetSize != nil {
		return ast.NewErr(s.Line, "max ret size given, but ret not defined")
	}

	return
}
