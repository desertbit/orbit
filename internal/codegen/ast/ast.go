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
	"time"

	"github.com/desertbit/orbit/internal/codegen/lexer"
	"github.com/desertbit/orbit/internal/utils"
)

// TODO: Decide whether a single lexer.Pos is enough, or if start and end should be recorded.

type File struct {
	Version *Version
	Srvc    *Service
	Types   []*Type
	Errs    []*Error
	Enums   []*Enum
}

type Version struct {
	Value int
	lexer.Pos
}

type Enum struct {
	Name   string
	Values []*EnumValue
	lexer.Pos
}

type EnumValue struct {
	Name  string
	Value int
	lexer.Pos
}

type Error struct {
	Name string
	ID   int
	lexer.Pos
}

type Type struct {
	Name   string
	Fields []*TypeField
	lexer.Pos
}

type TypeField struct {
	Name      string
	DataType  DataType
	StructTag string
	lexer.Pos
}

type Service struct {
	Calls   []*Call
	Streams []*Stream
	lexer.Pos
}

type Call struct {
	Name       string
	Arg        DataType
	Ret        DataType
	Async      bool
	Timeout    *time.Duration
	MaxArgSize *int64
	MaxRetSize *int64
	lexer.Pos
}

func (c *Call) NamePrv() string {
	return utils.FirstLower(c.Name)
}

type Stream struct {
	Name       string
	Arg        DataType
	Ret        DataType
	MaxArgSize *int64
	MaxRetSize *int64
	lexer.Pos
}

func (s *Stream) NamePrv() string {
	return utils.FirstLower(s.Name)
}
