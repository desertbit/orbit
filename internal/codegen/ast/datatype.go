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

import (
	"strings"
)

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
	// Returns go variable declaration.
	Decl() string
	// Returns go variable zero value.
	ZeroValue() string
	// Returns simple name.
	Name() string
}

type BaseType struct {
	DataType string
	Line     int
}

func (b *BaseType) Decl() string {
	if b.DataType == TypeTime {
		return "time.Time"
	}
	return b.DataType
}

func (b *BaseType) ZeroValue() string {
	switch b.DataType {
	case TypeByte, TypeUInt, TypeUInt8, TypeUInt16, TypeUInt32, TypeUInt64,
		TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64, TypeFloat32, TypeFloat64:
		return "0"
	case TypeString:
		return `""`
	case TypeBool:
		return "false"
	case TypeTime:
		return "time.Time{}"
	default:
		return "unknown base type zero value"
	}
}

func (b *BaseType) Name() string {
	if b.DataType == TypeTime {
		return "Time"
	}
	return strings.Title(b.DataType)
}

type MapType struct {
	Key   DataType
	Value DataType
	Line  int
}

func (m *MapType) Decl() string {
	return "map[" + m.Key.Decl() + "]" + m.Value.Decl()
}

func (m *MapType) ZeroValue() string {
	return "make(" + m.Decl() + ", 0)"
}

func (m *MapType) Name() string {
	return "Map" + strings.Title(m.Key.Name()) + strings.Title(m.Value.Name())
}

type ArrType struct {
	Elem DataType
	Line int
}

func (a *ArrType) Decl() string {
	return "[]" + a.Elem.Decl()
}

func (a *ArrType) ZeroValue() string {
	return "make(" + a.Decl() + ", 0)"
}

func (a *ArrType) Name() string {
	return "Arr" + strings.Title(a.Elem.Name())
}

type StructType struct {
	NamePrv string
	Line    int
}

func (s *StructType) Decl() string {
	return s.NamePrv
}

func (s *StructType) ZeroValue() string {
	return s.Decl() + "{}"
}

func (s *StructType) Name() string {
	return s.NamePrv
}

type EnumType struct {
	NamePrv string
	Line    int
}

func (e *EnumType) Decl() string {
	return e.NamePrv
}

func (e *EnumType) ZeroValue() string {
	return "0"
}

func (e *EnumType) Name() string {
	return e.NamePrv
}

type AnyType struct {
	NamePrv string
	Line    int
}

func (a *AnyType) Decl() string {
	return "unresolved any type"
}

func (a *AnyType) ZeroValue() string {
	return "unresolved any type"
}

func (a *AnyType) Name() string {
	return a.NamePrv
}
