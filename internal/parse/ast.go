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
	"strings"

	"github.com/desertbit/orbit/internal/utils"
)

type Error struct {
	Name string
	ID   int
}

type Service struct {
	Name    string
	Entries []Entry
}

type Entry interface {
	Args() *StructType
	HasArgs() bool

	Ret() *StructType
	HasRet() bool

	NamePub() string
	NamePrv() string
	Rev() bool
}

type Call struct {
	name string
	rev  bool
	args *StructType
	ret  *StructType

	Async bool
}

func (c *Call) Args() *StructType {
	return c.args
}

func (c *Call) HasArgs() bool {
	return c.args != nil
}

func (c *Call) Ret() *StructType {
	return c.ret
}

func (c *Call) HasRet() bool {
	return c.ret != nil
}

func (c *Call) NamePub() string {
	return strings.Title(c.name)
}

func (c *Call) NamePrv() string {
	return utils.ToLowerFirst(c.name)
}

func (c *Call) Rev() bool {
	return c.rev
}

type Stream struct {
	name string
	rev  bool
	args *StructType
	ret  *StructType
}

func (s *Stream) Args() *StructType {
	return s.args
}

func (s *Stream) HasArgs() bool {
	return s.args != nil
}

func (s *Stream) Ret() *StructType {
	return s.ret
}

func (s *Stream) HasRet() bool {
	return s.ret != nil
}

func (s *Stream) NamePub() string {
	return strings.Title(s.name)
}

func (s *Stream) NamePrv() string {
	return utils.ToLowerFirst(s.name)
}

func (s *Stream) Rev() bool {
	return s.rev
}

type Type struct {
	Name   string
	Fields []*TypeField

	serviceLocal bool
}

type TypeField struct {
	Name     string
	DataType DataType
}

type DataType interface {
	String() string
}

const (
	TypeByte   = "byte"
	TypeString = "string"
	TypeTime   = "time"

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
}

func (m *MapType) String() string {
	return "map[" + m.Key.String() + "]" + m.Value.String()
}

type ArrType struct {
	Elem DataType
}

func (a *ArrType) String() string {
	return "[]" + a.Elem.String()
}

type StructType struct {
	Name string
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
