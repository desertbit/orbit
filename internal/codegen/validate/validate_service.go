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

package validate

import (
	"errors"

	"github.com/desertbit/orbit/internal/codegen/ast"
)

func validateService(f *ast.File) (err error) {
	if f.Srvc == nil {
		return errors.New("no service definition found")
	}

	for j, c := range f.Srvc.Calls {
		for k := j + 1; k < len(f.Srvc.Calls); k++ {
			// Check for duplicate names.
			if c.Name == f.Srvc.Calls[k].Name {
				return ast.NewErr(c.Line, "call '%s' declared twice", c.Name)
			}
		}

		// Resolve the call.
		err = validateCall(c, f)
		if err != nil {
			return
		}
	}

	for j, s := range f.Srvc.Streams {
		for k := j + 1; k < len(f.Srvc.Streams); k++ {
			// Check for duplicate names.
			if s.Name == f.Srvc.Streams[k].Name {
				return ast.NewErr(s.Line, "stream '%s' declared twice", s.Name)
			}
		}

		// Resolve the stream.
		err = validateStream(s, f)
		if err != nil {
			return
		}
	}

	return
}

func validateCall(c *ast.Call, f *ast.File) (err error) {
	// Resolve all AnyTypes.
	c.Arg, err = resolveAnyType(c.Arg, f)
	if err != nil {
		return
	}
	c.Ret, err = resolveAnyType(c.Ret, f)
	if err != nil {
		return
	}

	// For arg and ret, only Struct types are allowed.
	if _, ok := c.Arg.(*ast.StructType); !ok && c.Arg != nil {
		return ast.NewErr(c.Line, "arg must be an inline or reference type")
	}
	if _, ok := c.Ret.(*ast.StructType); !ok && c.Ret != nil {
		return ast.NewErr(c.Line, "ret must be an inline or reference type")
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

func validateStream(s *ast.Stream, f *ast.File) (err error) {
	// Resolve all AnyTypes.
	s.Arg, err = resolveAnyType(s.Arg, f)
	if err != nil {
		return
	}
	s.Ret, err = resolveAnyType(s.Ret, f)
	if err != nil {
		return
	}

	// For arg and ret, only Struct types are allowed.
	if _, ok := s.Arg.(*ast.StructType); !ok && s.Arg != nil {
		return ast.NewErr(s.Line, "arg must be an inline or reference type")
	}
	if _, ok := s.Ret.(*ast.StructType); !ok && s.Ret != nil {
		return ast.NewErr(s.Line, "ret must be an inline or reference type")
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
