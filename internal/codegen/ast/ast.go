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

	"github.com/desertbit/orbit/internal/utils"
)

type Tree struct {
	Version int
	Srvc    *Service
	Types   []*Type
	Errs    []*Error
	Enums   []*Enum
}

type Enum struct {
	Name   string
	Values []*EnumValue
	Line   int
}

type EnumValue struct {
	Name  string
	Value int
	Line  int
}

type Error struct {
	Name string
	ID   int
	Line int
}

type Type struct {
	Name   string
	Fields []*TypeField
	Line   int
}

type TypeField struct {
	Name     string
	DataType DataType
	ValTag   string
	Line     int
}

type Service struct {
	Url     string
	Calls   []*Call
	Streams []*Stream
	Line    int
}

type Call struct {
	Name      string
	Arg       DataType
	ArgValTag string
	Ret       DataType
	RetValTag string
	Async     bool
	Timeout   *time.Duration
	Line      int
}

func (c *Call) NamePrv() string {
	return utils.NoTitle(c.Name)
}

type Stream struct {
	Name      string
	Arg       DataType
	ArgValTag string
	Ret       DataType
	RetValTag string
	Line      int
}

func (s *Stream) NamePrv() string {
	return utils.NoTitle(s.Name)
}
