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
	"github.com/desertbit/orbit/internal/codegen/lexer"
	"github.com/desertbit/orbit/internal/strutil"
)

const (
	TypeByte     = "byte"
	TypeString   = "string"
	TypeTime     = "time"
	TypeDuration = "duration"

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
	// Returns identifier.
	ID() string
	// Pos returns the lexer position.
	Pos() lexer.Pos
}

type dataType struct {
	pos     lexer.Pos
	pointer bool
}

func (d dataType) declBase() string {
	if d.pointer {
		return "*"
	} else {
		return ""
	}
}

func (d dataType) Pos() lexer.Pos {
	return d.pos
}

func (d dataType) Pointer() bool {
	return d.pointer
}

func newDataType(pos lexer.Pos, pointer bool) dataType {
	return dataType{
		pos:     pos,
		pointer: pointer,
	}
}

type BaseType struct {
	dataType

	DataType string
}

func NewBaseType(dataType string, pos lexer.Pos, pointer bool) *BaseType {
	return &BaseType{
		DataType: dataType,
		dataType: newDataType(pos, pointer),
	}
}

func (b *BaseType) Decl() string {
	d := b.declBase()
	switch b.DataType {
	case TypeTime:
		d += "time.Time"
	case TypeDuration:
		d += "time.Duration"
	default:
		d += b.DataType
	}
	return d
}

func (b *BaseType) ID() string {
	return b.DataType
}

type MapType struct {
	dataType

	Key   DataType
	Value DataType
}

func NewMapType(key, value DataType, pos lexer.Pos, pointer bool) *MapType {
	return &MapType{
		dataType: newDataType(pos, pointer),
		Key:      key,
		Value:    value,
	}
}

func (m *MapType) Decl() string {
	return m.declBase() + "map[" + m.Key.Decl() + "]" + m.Value.Decl()
}

func (m *MapType) ID() string {
	return m.Decl()
}

type ArrType struct {
	dataType

	Elem DataType
}

func NewArrType(elem DataType, pos lexer.Pos, pointer bool) *ArrType {
	return &ArrType{
		dataType: newDataType(pos, pointer),
		Elem:     elem,
	}
}

func (a *ArrType) Decl() string {
	return a.declBase() + "[]" + a.Elem.Decl()
}

func (a *ArrType) ID() string {
	return a.Decl()
}

type StructType struct {
	dataType

	Name string
}

func NewStructType(name string, pos lexer.Pos, pointer bool) *StructType {
	return &StructType{
		dataType: newDataType(pos, pointer),
		Name:     name,
	}
}

func (s *StructType) Decl() string {
	return s.declBase() + strutil.FirstUpper(s.Name)
}

func (s *StructType) ID() string {
	return s.Name
}

type EnumType struct {
	dataType

	Name string
}

func NewEnumType(name string, pos lexer.Pos, pointer bool) *EnumType {
	return &EnumType{
		dataType: newDataType(pos, pointer),
		Name:     name,
	}
}

func (e *EnumType) Decl() string {
	return e.declBase() + strutil.FirstUpper(e.Name)
}

func (e *EnumType) ID() string {
	return e.Name
}

type AnyType struct {
	dataType

	Name string
}

func NewAnyType(name string, pos lexer.Pos, pointer bool) *AnyType {
	return &AnyType{
		dataType: newDataType(pos, pointer),
		Name:     name,
	}
}

func (a *AnyType) Decl() string {
	return "unresolved any type" + a.declBase()
}

func (a *AnyType) ID() string {
	return a.Name
}
