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

package resolver

import (
	"github.com/desertbit/orbit/internal/codegen/ast"
)

func Resolve(
	srvcs []*ast.Service,
	types []*ast.Type,
	errs []*ast.Error,
	enums []*ast.Enum,
) (err error) {
	// Resolve all any types with their respective enum.
	for _, srvc := range srvcs {
		for _, c := range srvc.Calls {
			c.Args, err = resolveAnyType(c.Args, types, enums)
			if err != nil {
				return
			}
		}

		for _, s := range srvc.Streams {
			s.Args, err = resolveAnyType(s.Args, types, enums)
			if err != nil {
				return
			}
		}
	}

	for _, t := range types {
		for _, tf := range t.Fields {
			tf.DataType, err = resolveAnyType(tf.DataType, types, enums)
			if err != nil {
				return
			}
		}
	}

	return
}

func resolveAnyType(dt ast.DataType, types []*ast.Type, enums []*ast.Enum) (ast.DataType, error) {
	var err error

	switch v := dt.(type) {
	case *ast.AnyType:
		// Resolve the type.
		for _, t := range types {
			if t.Name == v.Name {
				return &ast.StructType{Name: v.Name, Line: v.Line}, nil
			}
		}
		for _, en := range enums {
			if en.Name == v.Name {
				return &ast.EnumType{Name: v.Name, Line: v.Line}, nil
			}
		}
		return nil, ast.NewErr(v.Line, "could not resolve type '%s'", v.Name)
	case *ast.MapType:
		v.Key, err = resolveAnyType(v.Key, types, enums)
		if err != nil {
			return nil, err
		}

		v.Value, err = resolveAnyType(v.Value, types, enums)
		if err != nil {
			return nil, err
		}
	case *ast.ArrType:
		v.Elem, err = resolveAnyType(v.Elem, types, enums)
		if err != nil {
			return nil, err
		}
	}

	// No change.
	return dt, nil
}
