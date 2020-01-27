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

package ast

const (
	TypeByte   = "byte"
	TypeString = "string"
	TypeTime   = "time"

	TypeBool = "bool"

	TypeInt   = "int"
	TypeInt8  = "int8"
	TypeInt16 = "int16"
	TypeInt32 = "int32"
	TypeInt64 = "int64"

	TypeUInt   = "uint"
	TypeUInt8  = "uint8"
	TypeUInt16 = "uint16"
	TypeUInt32 = "uint32"
	TypeUInt64 = "uint64"

	TypeFloat32 = "float32"
	TypeFloat64 = "float64"
)

type DataType interface {
	// Returns go string representation.
	String() string
}

type BaseType struct {
	DataType string
	Line     int
}

func (b *BaseType) String() string {
	if b.DataType == TypeTime {
		return "time.Time"
	}
	return b.DataType
}

type MapType struct {
	Key   DataType
	Value DataType
	Line  int
}

func (m *MapType) String() string {
	return "map[" + m.Key.String() + "]" + m.Value.String()
}

type ArrType struct {
	Elem DataType
	Line int
}

func (a *ArrType) String() string {
	return "[]" + a.Elem.String()
}

type StructType struct {
	Name string
	Line int
}

func (s *StructType) String() string {
	return "*" + s.Name
}

type EnumType struct {
	Name string
	Line int
}

func (e *EnumType) String() string {
	return e.Name
}

type AnyType struct {
	Name string
	Line int
}

func (a *AnyType) String() string {
	return "unresolved any type"
}
