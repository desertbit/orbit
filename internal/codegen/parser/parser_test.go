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

package parser_test

import (
	"os"
	"testing"
	"time"

	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/lexer"
	"github.com/desertbit/orbit/internal/codegen/parser"
	r "github.com/stretchr/testify/require"
)

var (
	c2Timeout          = 1 * time.Minute
	c2MaxArgSize int64 = 154 * 1024
	c2MaxRetSize int64 = 5 * 1024 * 1024
)

var (
	expVersion = &ast.Version{Value: 1, Pos: lexer.Pos{Line: 1, Column: 1}}
	c1         = &ast.Call{
		Name: "c1",
		Arg:  &ast.StructType{Name: "c1Arg"},
		Ret:  &ast.StructType{Name: "c1Ret"},
		Errors: []*ast.Error{
			{Name: "theFirstError", Pos: lexer.Pos{Line: 11, Column: 15}},
		},
	}
	c2 = &ast.Call{
		Name:       "c2",
		Async:      true,
		Arg:        &ast.StructType{Name: "c2Arg"},
		Ret:        &ast.StructType{Name: "c2Ret"},
		Timeout:    &c2Timeout,
		MaxArgSize: &c2MaxArgSize,
		MaxRetSize: &c2MaxRetSize,
		Errors: []*ast.Error{
			{Name: "theFirstError", Pos: lexer.Pos{Line: 24, Column: 15}},
			{Name: "theThirdError", Pos: lexer.Pos{Line: 24, Column: 30}},
		},
	}
	c3  = &ast.Call{Name: "c3"}
	rc1 = &ast.Call{
		Name: "rc1",
		Arg:  &ast.AnyType{Name: "Arg"},
		Ret:  &ast.StructType{Name: "Rc1Ret"},
	}
	rc2 = &ast.Call{
		Name:  "rc2",
		Async: true,
		Arg:   &ast.StructType{Name: "rc2Arg"},
	}
	rc3 = &ast.Call{
		Name: "rc3",
	}
	st1 = &ast.Stream{Name: "s1"}
	st2 = &ast.Stream{
		Name: "s2",
		Arg:  &ast.StructType{Name: "s2Arg"},
	}
	st3 = &ast.Stream{
		Name: "s3",
		Ret:  &ast.AnyType{Name: "Ret"},
	}
	rst1 = &ast.Stream{
		Name: "rs1",
		Arg:  &ast.AnyType{Name: "Arg"},
		Ret:  &ast.AnyType{Name: "Ret"},
	}
	rst2 = &ast.Stream{
		Name: "rs2",
	}
	expSrvc = &ast.Service{
		Calls:   []*ast.Call{c1, c2, c3, rc1, rc2, rc3},
		Streams: []*ast.Stream{st1, st2, st3, rst1, rst2},
	}

	expTypes = []*ast.Type{
		{
			Name: "c1Arg",
			Fields: []*ast.TypeField{
				{Name: "id", DataType: &ast.AnyType{Name: "int"}, StructTag: "json:\"ID\" yaml:\"id\""},
			},
		},
		{
			Name: "c1Ret",
			Fields: []*ast.TypeField{
				{Name: "sum", DataType: &ast.AnyType{Name: "float32"}},
			},
		},
		{
			Name: "c2Arg",
			Fields: []*ast.TypeField{
				{Name: "ts", DataType: &ast.AnyType{Name: "time"}},
			},
		},
		{
			Name: "c2Ret",
			Fields: []*ast.TypeField{
				{
					Name: "data",
					// []map[string][]Ret
					DataType: &ast.ArrType{
						Elem: &ast.MapType{
							Key:   &ast.AnyType{Name: "string"},
							Value: &ast.ArrType{Elem: &ast.AnyType{Name: "Ret"}},
						},
					},
				},
			},
		},
		{
			Name: "rc1Ret",
			Fields: []*ast.TypeField{
				{Name: "s", DataType: &ast.AnyType{Name: "string"}},
				{Name: "i", DataType: ast.NewAnyType("int", lexer.Pos{}, true)},
				{
					Name: "m",
					DataType: ast.NewMapType(
						&ast.AnyType{Name: "string"},
						&ast.AnyType{Name: "int"},
						lexer.Pos{},
						true,
					),
				},
				{Name: "sl", DataType: &ast.ArrType{Elem: &ast.AnyType{Name: "time"}}},
				{Name: "st", DataType: &ast.AnyType{Name: "Ret"}},
				{
					Name: "crazy",
					DataType: &ast.MapType{
						Key: &ast.AnyType{Name: "string"},
						Value: &ast.ArrType{
							Elem: &ast.ArrType{
								Elem: &ast.MapType{
									Key:   &ast.AnyType{Name: "string"},
									Value: &ast.AnyType{Name: "En1"},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "rc2Arg",
			Fields: []*ast.TypeField{
				{Name: "f", DataType: &ast.AnyType{Name: "float64"}},
				{Name: "b", DataType: &ast.AnyType{Name: "byte"}},
				{Name: "u8", DataType: &ast.AnyType{Name: "uint8"}},
				{Name: "u16", DataType: &ast.AnyType{Name: "uint16"}},
				{Name: "u32", DataType: &ast.AnyType{Name: "uint32"}},
				{Name: "u64", DataType: ast.NewAnyType("uint64", lexer.Pos{}, true)},
			},
		},
		{
			Name: "s2Arg",
			Fields: []*ast.TypeField{
				{Name: "id", DataType: &ast.AnyType{Name: "string"}, StructTag: "validator:\"required\""},
			},
		},
		{
			Name: "Arg",
			Fields: []*ast.TypeField{
				{Name: "s", DataType: &ast.AnyType{Name: "string"}, StructTag: "json:\"STRING\""},
				{Name: "i", DataType: &ast.AnyType{Name: "int"}},
				{
					Name: "m",
					DataType: &ast.MapType{
						Key:   &ast.AnyType{Name: "string"},
						Value: ast.NewAnyType("int", lexer.Pos{}, true),
					},
				},
				{Name: "sl", DataType: &ast.ArrType{Elem: &ast.AnyType{Name: "time"}}},
				{Name: "dur", DataType: &ast.AnyType{Name: "duration"}},
				{Name: "st", DataType: &ast.AnyType{Name: "Ret"}},
				{Name: "stp", DataType: ast.NewAnyType("Ret", lexer.Pos{}, true)},
				{
					Name: "crazy",
					DataType: &ast.MapType{
						Key: &ast.AnyType{Name: "string"},
						Value: &ast.ArrType{
							Elem: &ast.ArrType{
								Elem: &ast.MapType{
									Key:   &ast.AnyType{Name: "string"},
									Value: &ast.AnyType{Name: "En1"},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "Ret",
			Fields: []*ast.TypeField{
				{Name: "f", DataType: &ast.AnyType{Name: "float64"}},
				{Name: "b", DataType: &ast.AnyType{Name: "byte"}},
				{Name: "u8", DataType: &ast.AnyType{Name: "uint8"}},
				{Name: "u16", DataType: &ast.AnyType{Name: "uint16"}},
				{Name: "u32", DataType: &ast.AnyType{Name: "uint32"}},
				{Name: "u64", DataType: &ast.AnyType{Name: "uint64"}},
			},
		},
	}

	expEnums = []*ast.Enum{
		{
			Name: "En1",
			Values: []*ast.EnumValue{
				{Name: "Val1", Value: 1},
				{Name: "Val2", Value: 2},
				{Name: "Val3", Value: 3},
			},
		},
	}

	expErrs = []*ast.Error{
		{Name: "theFirstError", ID: 1},
		{Name: "theSecondError", ID: 2},
		{Name: "theThirdError", ID: 3},
	}
)

func TestParser_Parse(t *testing.T) {
	t.Run("parseValid", testParseValid)
}

func testParseValid(t *testing.T) {
	t.Parallel()

	// Read valid .orbit file from testdata.
	input, err := os.ReadFile("./testdata/valid.orbit")
	r.NoError(t, err)

	// Parse file.
	f, err := parser.Parse(lexer.Lex(string(input)))
	r.NoError(t, err)

	// Version.
	r.Exactly(t, expVersion, f.Version)

	// Services.
	r.NotNil(t, f.Srvc)
	requireEqualService(t, expSrvc, f.Srvc)

	// Types.
	r.Len(t, f.Types, len(expTypes))
	for i, expType := range expTypes {
		requireEqualType(t, expType, f.Types[i])
	}

	// Enums.
	r.Len(t, f.Enums, len(expEnums))
	for i, expEn := range expEnums {
		requireEqualEnum(t, expEn, f.Enums[i])
	}

	// Errors.
	r.Len(t, f.Errs, len(expErrs))
	for i, expErr := range expErrs {
		requireEqualError(t, expErr, f.Errs[i])
	}
}

//###############//
//### Helpers ###//
//###############//

func requireEqualService(t *testing.T, exp, act *ast.Service) {
	r.Len(t, act.Calls, len(exp.Calls))
	r.Len(t, act.Streams, len(exp.Streams))
	for i, expc := range exp.Calls {
		requireEqualCall(t, expc, act.Calls[i])
	}
	for i, exps := range exp.Streams {
		requireEqualStream(t, exps, act.Streams[i])
	}
}

func requireEqualCall(t *testing.T, exp, act *ast.Call) {
	r.Exactly(t, exp.Name, act.Name)
	r.Exactly(t, exp.Async, act.Async)
	r.Exactly(t, exp.Timeout, act.Timeout)
	r.Exactly(t, exp.MaxArgSize, act.MaxArgSize)
	r.Exactly(t, exp.MaxRetSize, act.MaxRetSize)
	r.Exactly(t, exp.Errors, act.Errors)
	requireEqualDataType(t, exp.Arg, act.Arg)
	requireEqualDataType(t, exp.Ret, act.Ret)
}

func requireEqualStream(t *testing.T, exp, act *ast.Stream) {
	r.Exactly(t, exp.Name, act.Name)
	r.Exactly(t, exp.MaxArgSize, act.MaxArgSize)
	r.Exactly(t, exp.MaxRetSize, act.MaxRetSize)
	requireEqualDataType(t, exp.Arg, act.Arg)
	requireEqualDataType(t, exp.Ret, act.Ret)
}

func requireEqualType(t *testing.T, exp, act *ast.Type) {
	r.Exactly(t, exp.Name, act.Name)
	r.Len(t, exp.Fields, len(act.Fields))
	for i, exptf := range exp.Fields {
		r.Exactly(t, exptf.Name, act.Fields[i].Name)
		r.Exactly(t, exptf.StructTag, act.Fields[i].StructTag)
		requireEqualDataType(t, exptf.DataType, act.Fields[i].DataType)
	}
}

func requireEqualEnum(t *testing.T, exp, act *ast.Enum) {
	r.Exactly(t, exp.Name, act.Name)
	r.Len(t, exp.Values, len(act.Values))
	for i, expv := range exp.Values {
		r.Exactly(t, expv.Name, act.Values[i].Name)
		r.Exactly(t, expv.Value, act.Values[i].Value)
	}
}

func requireEqualError(t *testing.T, exp, act *ast.Error) {
	r.Exactly(t, exp.Name, act.Name)
	r.Exactly(t, exp.ID, act.ID)
}

func requireEqualDataType(t *testing.T, exp, act ast.DataType) {
	if exp == nil && act == nil {
		return
	}

	r.IsType(t, exp, act)
	switch v := exp.(type) {
	case *ast.StructType, *ast.EnumType, *ast.AnyType:
		r.Exactly(t, exp.Decl(), act.Decl())
	case *ast.ArrType:
		requireEqualDataType(t, v.Elem, act.(*ast.ArrType).Elem)
	case *ast.MapType:
		requireEqualDataType(t, v.Key, act.(*ast.MapType).Key)
		requireEqualDataType(t, v.Value, act.(*ast.MapType).Value)
	default:
		t.Fatalf("unexpected data type %v", v)
	}
}
