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

/*type EntryType int

const (
	EntryTypeCall    EntryType = iota
	EntryTypeRevCall EntryType = iota
	EntryTypeStream  EntryType = iota
)*/

type Entry interface {
	Name() string
}

type EntryParam struct {
	Type     *StructType
	IsStream bool
}

type Call struct {
	name string

	Args *EntryParam
	Ret  *EntryParam
}

func (c *Call) Name() string {
	return c.name
}

type RevCall struct {
	name string

	Args *EntryParam
	Ret  *EntryParam
}

func (r *RevCall) Name() string {
	return r.name
}

type Stream struct {
	name string
}

func (s *Stream) Name() string {
	return s.name
}

type Type interface {
	Name() string
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

type StructType struct {
	name string

	Fields []Type
}

func (s *StructType) Name() string {
	return s.name
}

type BaseType struct {
	name     string
	dataType string
}

func (b *BaseType) Name() string {
	return b.name
}

func (b *BaseType) DataType() string {
	return b.dataType
}

type MapType struct {
	name string

	Key   Type
	Value Type
}

func (m *MapType) Name() string {
	return m.name
}

type ArrType struct {
	name string

	ElemType Type
}

func (a *ArrType) Name() string {
	return a.name
}
