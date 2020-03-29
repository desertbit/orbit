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
	"errors"

	"github.com/desertbit/orbit/internal/codegen/ast"
)

// Resolve resolves the given tree by replacing any AnyTypes with their respective
// EnumType or StructType.
// In addition, a validation is performed, like checking for duplicate names, etc.
func Resolve(tree *ast.Tree) (err error) {
	// Service.
	if tree.Srvc == nil {
		return errors.New("no service definition found")
	}

	err = resolveService(tree.Srvc, tree.Types, tree.Enums)
	if err != nil {
		return
	}

	// Types.
	for i, t := range tree.Types {
		for j := i + 1; j < len(tree.Types); j++ {
			// Check for duplicate names.
			if t.Name == tree.Types[j].Name {
				return ast.NewErr(t.Line, "type '%s' declared twice", t.Name)
			}
		}

		for j, tf := range t.Fields {
			for k := j + 1; k < len(t.Fields); k++ {
				// Check for duplicate field names.
				if tf.Name == t.Fields[k].Name {
					return ast.NewErr(
						tf.Line, "field '%s' of type '%s' declared twice", tf.Name, t.Name,
					)
				}
			}

			// Resolve all AnyTypes.
			tf.DataType, err = resolveAnyType(tf.DataType, tree.Types, tree.Enums)
			if err != nil {
				return
			}
		}
	}

	// Errors.
	for i, e := range tree.Errs {
		for j := i + 1; j < len(tree.Errs); j++ {
			e2 := tree.Errs[j]

			// Check for duplicate names.
			if e.Name == e2.Name {
				return ast.NewErr(e.Line, "error '%s' declared twice", e.Name)
			}

			// Check for valid id.
			if e.ID <= 0 {
				return ast.NewErr(e.Line, "invalid error id, must be greater than 0")
			}

			// Check for duplicate id.
			if e.ID == e2.ID {
				return ast.NewErr(e.Line, "error '%s' has same id as '%s'", e.Name, e2.Name)
			}
		}
	}

	// Enums.
	for i, en := range tree.Enums {
		for j := i + 1; j < len(tree.Enums); j++ {
			// Check for duplicate name.
			if en.Name == tree.Enums[j].Name {
				return ast.NewErr(en.Line, "enum '%s' declared twice", en.Name)
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
			if t.Name == v.Name() {
				return &ast.StructType{NamePrv: v.Name(), Line: v.Line}, nil
			}
		}
		for _, en := range enums {
			if en.Name == v.Name() {
				return &ast.EnumType{NamePrv: v.Name(), Line: v.Line}, nil
			}
		}
		return nil, ast.NewErr(v.Line, "resolve: unknown type '%s'", v.Name())
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
