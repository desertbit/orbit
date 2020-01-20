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

package parse

import (
	"github.com/desertbit/orbit/internal/utils"
)

type Error struct {
	Name string
	ID   int

	line int
}

type Type struct {
	Name   string
	Fields []*TypeField

	line int
}

type TypeField struct {
	Name     string
	DataType DataType

	line int
}

type Service struct {
	Name    string
	Calls   []*Call
	Streams []*Stream
	Errors  []*Error
	Types   []*Type

	line int
}

type Call struct {
	Name  string
	Rev   bool
	Args  *StructType
	Ret   *StructType
	Async bool

	line int
}

func (c *Call) NamePrv() string {
	return utils.ToLowerFirst(c.Name)
}

type Stream struct {
	Name string
	Rev  bool
	Args *StructType
	Ret  *StructType

	line int
}

func (s *Stream) NamePrv() string {
	return utils.ToLowerFirst(s.Name)
}

type DataType interface {
	String() string
}

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

type BaseType struct {
	dataType string
	line     int
}

func (b *BaseType) String() string {
	if b.dataType == TypeTime {
		return "time.Time"
	}
	return b.dataType
}

type MapType struct {
	Key   DataType
	Value DataType

	line int
}

func (m *MapType) String() string {
	return "map[" + m.Key.String() + "]" + m.Value.String()
}

type ArrType struct {
	Elem DataType

	line int
}

func (a *ArrType) String() string {
	return "[]" + a.Elem.String()
}

type StructType struct {
	Name string

	line int
}

func (s *StructType) String() string {
	return "*" + s.Name
}

func containsStruct(dt DataType) (s *StructType) {
	switch v := dt.(type) {
	case *BaseType:
		return nil
	case *StructType:
		return v
	case *MapType:
		return containsStruct(v.Value)
	case *ArrType:
		return containsStruct(v.Elem)
	default:
		return nil
	}
}
