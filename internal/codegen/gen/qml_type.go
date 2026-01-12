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
	"os"
	"path/filepath"
	"slices"

	"github.com/desertbit/orbit/internal/codegen/ast"
)

func (g *qmlGenerator) genTypes(ts []*ast.Type, skipTypeNames []string) (err error) {
	for _, t := range ts {
		if slices.Contains(skipTypeNames, t.Ident()) {
			continue
		}

		// Imports.
		g.writeLn("import QtQml")
		g.writeLn("")

		// Definition.
		g.writeLn("QtObject {")
		g.indent(func() {
			g.writeLn("id: root")
			g.writeLn("")

			// Generate the properties.
			for _, f := range t.Fields {
				qmlType, defaultValue := qmlDataType(f.DataType)
				qmlName := qmlSanitizeName(f.IdentPrv())

				g.writef("property %s %s", qmlType, qmlName)
				if defaultValue != "" {
					g.writef(": %s", defaultValue)
				}

				// Add a comment about the type inside a javascript array, since we can not use the qml list for now.
				if dt, ok := f.DataType.(*ast.ArrType); ok {
					g.writef(" // Contains %s objects.", dt.Elem.ID())
				}

				g.writeLn("")

				// Duration fields are special as in that nanoseconds are usually very unwieldy.
				// Generate a read only property that returns it as milliseconds.
				if bt, ok := f.DataType.(*ast.BaseType); ok && bt.DataType == ast.TypeDuration {
					g.writefLn("readonly property %[1]s %[2]sMs: root.%[2]s / 1000000", qmlType, qmlName)
				}
			}
			g.writeLn("")

			// Generate a load function to load data from a javascript object.
			g.writeLn("function load(o: var): void {")
			g.indent(func() {
				for _, f := range t.Fields {
					g.qmlGenLoad(f.IdentPrv(), f.Name, f.DataType)
				}
			})
			g.writeLn("}")
			g.writeLn("")

			// Generate a reset function to reset all properties to their default values.
			g.writeLn("function reset(): void {")
			g.indent(func() {
				for _, f := range t.Fields {
					g.qmlGenReset(f.IdentPrv(), f.DataType)
				}
			})
			g.writeLn("}")
		})
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

func qmlDataType(dt ast.DataType) (qmlType, defaultValue string) {
	switch v := dt.(type) {
	case *ast.StructType:
		qmlType = v.ID()
		defaultValue = v.ID() + "{}"

	case *ast.ArrType:
		qmlType = "var"
		defaultValue = "[]"

	case *ast.MapType:
		qmlType = "var"
		defaultValue = "({})"

	case *ast.BaseType:
		switch v.DataType {
		case ast.TypeString:
			qmlType = "string"
			defaultValue = `""`
		case ast.TypeTime:
			qmlType = "date"
			defaultValue = "new Date()"
		case ast.TypeBool:
			qmlType = "bool"
			defaultValue = "false"
		case ast.TypeInt, ast.TypeInt8, ast.TypeInt16, ast.TypeInt32, ast.TypeInt64,
			ast.TypeUInt, ast.TypeUInt8, ast.TypeUInt16, ast.TypeUInt32, ast.TypeUInt64,
			ast.TypeByte,
			ast.TypeDuration:
			qmlType = "int"
			defaultValue = "0"
		case ast.TypeFloat32:
			qmlType = "real"
			defaultValue = "0"
		case ast.TypeFloat64:
			qmlType = "double"
			defaultValue = "0"
		}

	case *ast.EnumType:
		qmlType = "int"
		defaultValue = "0"
	}

	return
}

// qmlGenLoad generates the load statement for the field with the given name and data type.
// qmlName is the name of the QML property, loadName the name of the property to be loaded.
func (g *qmlGenerator) qmlGenLoad(qmlName, loadName string, dt ast.DataType) {
	qmlName = qmlSanitizeName(qmlName)

	switch v := dt.(type) {
	case *ast.StructType:
		g.writefLn("root.%s.load(o.%s)", qmlName, loadName)

	/*case *ast.ArrType:
	g.writefLn("root.%s.length = o.%s.length", qmlName, loadName)
	g.writefLn("for (let i = 0; i < root.%s.length; ++i) {", qmlName)
	g.indent(func() {
		g.qmlGenLoad(qmlName+"[i]", loadName+"[i]", v.Elem)
	})
	g.writeLn("}")*/

	case *ast.BaseType:
		if v.DataType == ast.TypeTime {
			g.writefLn("root.%[1]s = new Date(o.%[2]s === null ? 'foobar' : Date.parse(o.%[2]s))", qmlName, loadName)
		} else {
			g.writefLn("root.%s = o.%s", qmlName, loadName)
		}

	default:
		g.writefLn("root.%s = o.%s", qmlName, loadName)
	}
}

// qmlGenReset generates the reset statement for the field with the given name and data type.
// qmlName is the name of the QML property, loadName the name of the property to be loaded.
func (g *qmlGenerator) qmlGenReset(qmlName string, dt ast.DataType) {
	qmlName = qmlSanitizeName(qmlName)

	switch dt.(type) {
	case *ast.StructType:
		g.writefLn("root.%s.reset()", qmlName)

	case *ast.MapType:
		g.writefLn("root.%s = {}", qmlName)

	/*case *ast.ArrType:
	g.writefLn("root.%s.length = 0", qmlName)*/

	default:
		_, defaultValue := qmlDataType(dt)
		g.writefLn("root.%s = %s", qmlName, defaultValue)
	}
}

// qmlSanitizeName ensures the given name is in conformance with the QML naming conventions.
func qmlSanitizeName(name string) string {
	switch name {
	case "id":
		// The id keyword is very special in QML and should be avoided.
		return "mid"

	case "default":
		// default is a reserved keyword.
		return "mdefault"

	default:
		return name
	}
}
