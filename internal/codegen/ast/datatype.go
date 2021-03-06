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
}

type BaseType struct {
	DataType string
	lexer.Pos
}

func (b *BaseType) Decl() string {
	if b.DataType == TypeTime {
		return "time.Time"
	}
	if b.DataType == TypeDuration {
		return "time.Duration"
	}
	return b.DataType
}

func (b *BaseType) ID() string {
	return b.DataType
}

type MapType struct {
	Key   DataType
	Value DataType
	lexer.Pos
}

func (m *MapType) Decl() string {
	return "map[" + m.Key.Decl() + "]" + m.Value.Decl()
}

func (m *MapType) ID() string {
	return m.Decl()
}

type ArrType struct {
	Elem DataType
	lexer.Pos
}

func (a *ArrType) Decl() string {
	return "[]" + a.Elem.Decl()
}

func (a *ArrType) ID() string {
	return a.Decl()
}

type StructType struct {
	Name string
	lexer.Pos
}

func (s *StructType) Decl() string {
	return strutil.FirstUpper(s.Name)
}

func (s *StructType) ID() string {
	return s.Name
}

type EnumType struct {
	Name string
	lexer.Pos
}

func (e *EnumType) Decl() string {
	return strutil.FirstUpper(e.Name)
}

func (e *EnumType) ID() string {
	return e.Name
}

type AnyType struct {
	Name string
	lexer.Pos
}

func (a *AnyType) Decl() string {
	return "unresolved any type"
}

func (a *AnyType) ID() string {
	return a.Name
}
