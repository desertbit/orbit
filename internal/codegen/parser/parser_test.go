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
	"io/ioutil"
	"testing"
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/parser"
	"github.com/stretchr/testify/require"
)

var (
	version            = 1
	c2Timeout          = 500 * time.Millisecond
	c2MaxArgSize int64 = 154 * 1024
	c2MaxRetSize int64 = 5 * 1024 * 1024
)

var (
	c1 = &ast.Call{
		Name: "C1",
		Arg:  &ast.StructType{NamePrv: "C1Arg"},
		Ret:  &ast.StructType{NamePrv: "C1Ret"},
	}
	c2 = &ast.Call{
		Name:       "C2",
		Async:      true,
		Arg:        &ast.StructType{NamePrv: "C2Arg"},
		Ret:        &ast.StructType{NamePrv: "C2Ret"},
		Timeout:    &c2Timeout,
		MaxArgSize: &c2MaxArgSize,
		MaxRetSize: &c2MaxRetSize,
	}
	c3  = &ast.Call{Name: "C3"}
	rc1 = &ast.Call{
		Name: "Rc1",
		Arg:  &ast.AnyType{NamePrv: "Arg"},
		Ret:  &ast.StructType{NamePrv: "Rc1Ret"},
	}
	rc2 = &ast.Call{
		Name:  "Rc2",
		Async: true,
		Arg:   &ast.StructType{NamePrv: "Rc2Arg"},
	}
	rc3 = &ast.Call{
		Name: "Rc3",
	}
	st1 = &ast.Stream{Name: "S1"}
	st2 = &ast.Stream{
		Name: "S2",
		Arg:  &ast.StructType{NamePrv: "S2Arg"},
	}
	st3 = &ast.Stream{
		Name: "S3",
		Ret:  &ast.AnyType{NamePrv: "En1"},
	}
	rst1 = &ast.Stream{
		Name: "Rs1",
		Arg:  &ast.AnyType{NamePrv: "Arg"},
		Ret:  &ast.AnyType{NamePrv: "Ret"},
	}
	rst2 = &ast.Stream{
		Name: "Rs2",
	}
	expSrvc = &ast.Service{
		Calls:   []*ast.Call{c1, c2, c3, rc1, rc2, rc3},
		Streams: []*ast.Stream{st1, st2, st3, rst1, rst2},
	}

	expTypes = []*ast.Type{
		{
			Name: "C1Arg",
			Fields: []*ast.TypeField{
				{Name: "Id", DataType: &ast.AnyType{NamePrv: "int"}, StructTag: "json:\"ID\" yaml:\"id\""},
			},
		},
		{
			Name: "C1Ret",
			Fields: []*ast.TypeField{
				{Name: "Sum", DataType: &ast.AnyType{NamePrv: "float32"}},
			},
		},
		{
			Name: "C2Arg",
			Fields: []*ast.TypeField{
				{Name: "Ts", DataType: &ast.AnyType{NamePrv: "time"}},
			},
		},
		{
			Name: "C2Ret",
			Fields: []*ast.TypeField{
				{
					Name: "Data",
					// []map[string][]Ret
					DataType: &ast.ArrType{
						Elem: &ast.MapType{
							Key:   &ast.AnyType{NamePrv: "string"},
							Value: &ast.ArrType{Elem: &ast.AnyType{NamePrv: "Ret"}},
						},
					},
				},
			},
		},
		{
			Name: "Rc1Ret",
			Fields: []*ast.TypeField{
				{Name: "S", DataType: &ast.AnyType{NamePrv: "string"}},
				{Name: "I", DataType: &ast.AnyType{NamePrv: "int"}},
				{
					Name: "M",
					DataType: &ast.MapType{
						Key:   &ast.AnyType{NamePrv: "string"},
						Value: &ast.AnyType{NamePrv: "int"},
					},
				},
				{Name: "Sl", DataType: &ast.ArrType{Elem: &ast.AnyType{NamePrv: "time"}}},
				{Name: "St", DataType: &ast.AnyType{NamePrv: "Ret"}},
				{
					Name: "Crazy",
					DataType: &ast.MapType{
						Key: &ast.AnyType{NamePrv: "string"},
						Value: &ast.ArrType{
							Elem: &ast.ArrType{
								Elem: &ast.MapType{
									Key:   &ast.AnyType{NamePrv: "string"},
									Value: &ast.AnyType{NamePrv: "En1"},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "Rc2Arg",
			Fields: []*ast.TypeField{
				{Name: "F", DataType: &ast.AnyType{NamePrv: "float64"}},
				{Name: "B", DataType: &ast.AnyType{NamePrv: "byte"}},
				{Name: "U8", DataType: &ast.AnyType{NamePrv: "uint8"}},
				{Name: "U16", DataType: &ast.AnyType{NamePrv: "uint16"}},
				{Name: "U32", DataType: &ast.AnyType{NamePrv: "uint32"}},
				{Name: "U64", DataType: &ast.AnyType{NamePrv: "uint64"}},
			},
		},
		{
			Name: "S2Arg",
			Fields: []*ast.TypeField{
				{Name: "Id", DataType: &ast.AnyType{NamePrv: "string"}, StructTag: "validator:\"required\""},
			},
		},
		{
			Name: "Arg",
			Fields: []*ast.TypeField{
				{Name: "S", DataType: &ast.AnyType{NamePrv: "string"}, StructTag: "json:\"STRING\""},
				{Name: "I", DataType: &ast.AnyType{NamePrv: "int"}},
				{
					Name: "M",
					DataType: &ast.MapType{
						Key:   &ast.AnyType{NamePrv: "string"},
						Value: &ast.AnyType{NamePrv: "int"},
					},
				},
				{Name: "Sl", DataType: &ast.ArrType{Elem: &ast.AnyType{NamePrv: "time"}}},
				{Name: "Dur", DataType: &ast.AnyType{NamePrv: "duration"}},
				{Name: "St", DataType: &ast.AnyType{NamePrv: "Ret"}},
				{
					Name: "Crazy",
					DataType: &ast.MapType{
						Key: &ast.AnyType{NamePrv: "string"},
						Value: &ast.ArrType{
							Elem: &ast.ArrType{
								Elem: &ast.MapType{
									Key:   &ast.AnyType{NamePrv: "string"},
									Value: &ast.AnyType{NamePrv: "En1"},
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
				{Name: "F", DataType: &ast.AnyType{NamePrv: "float64"}},
				{Name: "B", DataType: &ast.AnyType{NamePrv: "byte"}},
				{Name: "U8", DataType: &ast.AnyType{NamePrv: "uint8"}},
				{Name: "U16", DataType: &ast.AnyType{NamePrv: "uint16"}},
				{Name: "U32", DataType: &ast.AnyType{NamePrv: "uint32"}},
				{Name: "U64", DataType: &ast.AnyType{NamePrv: "uint64"}},
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
		{Name: "TheFirstError", ID: 1},
		{Name: "TheSecondError", ID: 2},
		{Name: "TheThirdError", ID: 3},
	}
)

func TestParser_Parse(t *testing.T) {
	t.Run("parseValid", testParseValid)
}

func testParseValid(t *testing.T) {
	t.Parallel()

	// Read valid .orbit file from testdata.
	input, err := ioutil.ReadFile("./testdata/valid.orbit")
	require.NoError(t, err)

	// Parse file.
	f, err := parser.Parse(closer.New(), string(input))
	require.NoError(t, err)

	// Version.
	require.Exactly(t, version, f.Version)

	// Services.
	require.NotNil(t, f.Srvc)
	requireEqualService(t, expSrvc, f.Srvc)

	// Types.
	require.Len(t, f.Types, len(expTypes))
	for i, expType := range expTypes {
		requireEqualType(t, expType, f.Types[i])
	}

	// Enums.
	require.Len(t, f.Enums, len(expEnums))
	for i, expEn := range expEnums {
		requireEqualEnum(t, expEn, f.Enums[i])
	}

	// Errors.
	require.Len(t, f.Errs, len(expErrs))
	for i, expErr := range expErrs {
		requireEqualError(t, expErr, f.Errs[i])
	}
}

//###############//
//### Helpers ###//
//###############//

func requireEqualService(t *testing.T, exp, act *ast.Service) {
	require.Len(t, act.Calls, len(exp.Calls))
	require.Len(t, act.Streams, len(exp.Streams))
	for i, expc := range exp.Calls {
		requireEqualCall(t, expc, act.Calls[i])
	}
	for i, exps := range exp.Streams {
		requireEqualStream(t, exps, act.Streams[i])
	}
}

func requireEqualCall(t *testing.T, exp, act *ast.Call) {
	require.Exactly(t, exp.Name, act.Name)
	require.Exactly(t, exp.Async, act.Async)
	require.Exactly(t, exp.Timeout, act.Timeout)
	require.Exactly(t, exp.MaxArgSize, act.MaxArgSize)
	require.Exactly(t, exp.MaxRetSize, act.MaxRetSize)
	requireEqualDataType(t, exp.Arg, act.Arg)
	requireEqualDataType(t, exp.Ret, act.Ret)
}

func requireEqualStream(t *testing.T, exp, act *ast.Stream) {
	require.Exactly(t, exp.Name, act.Name)
	require.Exactly(t, exp.MaxArgSize, act.MaxArgSize)
	require.Exactly(t, exp.MaxRetSize, act.MaxRetSize)
	requireEqualDataType(t, exp.Arg, act.Arg)
	requireEqualDataType(t, exp.Ret, act.Ret)
}

func requireEqualType(t *testing.T, exp, act *ast.Type) {
	require.Exactly(t, exp.Name, act.Name)
	require.Len(t, exp.Fields, len(act.Fields))
	for i, exptf := range exp.Fields {
		require.Exactly(t, exptf.Name, act.Fields[i].Name)
		require.Exactly(t, exptf.StructTag, act.Fields[i].StructTag)
		requireEqualDataType(t, exptf.DataType, act.Fields[i].DataType)
	}
}

func requireEqualEnum(t *testing.T, exp, act *ast.Enum) {
	require.Exactly(t, exp.Name, act.Name)
	require.Len(t, exp.Values, len(act.Values))
	for i, expv := range exp.Values {
		require.Exactly(t, expv.Name, act.Values[i].Name)
		require.Exactly(t, expv.Value, act.Values[i].Value)
	}
}

func requireEqualError(t *testing.T, exp, act *ast.Error) {
	require.Exactly(t, exp.Name, act.Name)
	require.Exactly(t, exp.ID, act.ID)
}

func requireEqualDataType(t *testing.T, exp, act ast.DataType) {
	if exp == nil && act == nil {
		return
	}

	require.IsType(t, exp, act)
	switch v := exp.(type) {
	case *ast.StructType, *ast.EnumType, *ast.AnyType:
		require.Exactly(t, exp.Decl(), act.Decl())
	case *ast.ArrType:
		requireEqualDataType(t, v.Elem, act.(*ast.ArrType).Elem)
	case *ast.MapType:
		requireEqualDataType(t, v.Key, act.(*ast.MapType).Key)
		requireEqualDataType(t, v.Value, act.(*ast.MapType).Value)
	default:
		t.Fatalf("unexpected data type %v", v)
	}
}
