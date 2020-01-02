/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2019 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2019 Sebastian Borchers <sebastian[at]desertbit.com>
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

type Service struct {
	Name    string
	Entries []Entry
}

type Entry interface {
	Name() string
	Rev() bool
}

type EntryParam struct {
	Type     *StructType
	IsStream bool
}

type Call struct {
	name string
	rev  bool

	Args *EntryParam
	Ret  *EntryParam
}

func (c *Call) Name() string {
	return c.name
}

func (c *Call) Rev() bool {
	return c.rev
}

type Stream struct {
	name string
	rev  bool
}

func (s *Stream) Name() string {
	return s.name
}

func (s *Stream) Rev() bool {
	return s.rev
}

type Type interface{}

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

type StructType struct {
	Name   string
	Fields []*StructField
}

type StructField struct {
	Name string
	Type Type
}

type BaseType struct {
	dataType string
}

func (b *BaseType) DataType() string {
	return b.dataType
}

type MapType struct {
	Key   Type
	Value Type
}

type ArrType struct {
	ElemType Type
}
