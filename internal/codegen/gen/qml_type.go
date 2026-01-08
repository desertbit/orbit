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
	"fmt"
	"os"
	"path/filepath"

	"github.com/desertbit/orbit/internal/codegen/ast"
)

func (g *qmlGenerator) genTypes(ts []*ast.Type) (err error) {
	for _, t := range ts {
		// Imports.
		g.writeLn("import QtQml")
		g.writeLn("")
		g.writeLn("import Lib as L")
		g.writeLn("")

		// Definition.
		g.writeLn("QtObject {")

		for _, f := range t.Fields {
			g.write("    property ")

			var defaultValue string
			switch v := f.DataType.(type) {
			case *ast.StructType:

			case *ast.ArrType:

			case *ast.MapType:

			case *ast.BaseType:
				switch v.DataType {
				case ast.TypeString:
					g.write("string")
					defaultValue = `""`
				case ast.TypeTime:
					g.write("date")
					defaultValue = "L.Date.Invalid"
				case ast.TypeBool:
					g.write("bool")
					defaultValue = "false"
				case ast.TypeInt, ast.TypeInt8, ast.TypeInt16, ast.TypeInt32, ast.TypeInt64,
					ast.TypeUInt, ast.TypeUInt8, ast.TypeUInt16, ast.TypeUInt32, ast.TypeUInt64,
					ast.TypeByte,
					ast.TypeDuration:
					g.write("int")
					defaultValue = "0"
				case ast.TypeFloat32:
					g.write("real")
					defaultValue = "0"
				case ast.TypeFloat64:
					g.write("double")
					defaultValue = "0"
				default:
					return fmt.Errorf("unknown base data type %s (field.Name=%s, type.Name=%s)", v.DataType, f.Name, t.Name)
				}

			case *ast.EnumType:

			default:
				return fmt.Errorf("unsupported data type %s (field.Name=%s, type.Name=%s)", f.DataType.ID(), f.Name, t.Name)
			}

			g.writef(" %s", f.Name)
			if defaultValue != "" {
				g.writef(": %s", defaultValue)
			}
			g.writeLn("")
		}

		g.writeLn("}")

		// Create the file.
		err = os.WriteFile(filepath.Join(g.dir, t.Ident()+".qml"), []byte(g.s.String()), 0o644)
		if err != nil {
			return
		}

		g.s.Reset()
	}

	return
}
