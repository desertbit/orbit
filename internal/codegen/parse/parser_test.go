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

package parse_test

import (
	"strings"
	"testing"

	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/parse"
	"github.com/desertbit/orbit/internal/codegen/token"
	"github.com/stretchr/testify/require"
)

const orbit = `
service {
    call c1 {
        arg: int
        ret: float32
    }
    call c2 {
        async
        arg: time
        ret: []map[string][]Ret
    }
    call c3 {}

    call rc1 {
        arg: Args
        ret: {
            s string
            i int
            m map[string]int
            sl []time
            st Ret
            crazy map[string][][]map[string]En1
        }
    }
    call rc2 {
        async
        arg: {
            f float64
            b byte
            u8 uint8
            u16 uint16
            u32 uint32
            u64 uint64
        }
    }
    call rc3 {}

    stream s1 {}
    stream s2 {
        arg: string
    }
    stream s3 {
        ret: En1
    }

    stream rs1 {
        arg: Args
        ret: Ret
    }
    stream rs2 {
        arg: map[string]int
    }
    stream rs3 {}
}

type Args {
    s string
    i int
    m map[string]int
    sl []time
    st Ret
    crazy map[string][][]map[string]En1
}

type Ret {
    f float64
    b byte
    u8 uint8
    u16 uint16
    u32 uint32
    u64 uint64
}

enum En1 {
    Val1 = 1
    Val2 = 2
    Val3 = 3
}

errors {
    theFirstError = 1
    theSecondError = 2
    theThirdError = 3
}
`

var (
	c1 = &ast.Call{
		Name: "C1",
		Arg:  &ast.BaseType{DataType: ast.TypeInt},
		Ret:  &ast.BaseType{DataType: ast.TypeFloat32},
	}
	c2 = &ast.Call{
		Name:  "C2",
		Async: true,
		Arg:   &ast.BaseType{DataType: ast.TypeTime},
		Ret: &ast.ArrType{
			Elem: &ast.MapType{
				Key:   &ast.BaseType{DataType: ast.TypeString},
				Value: &ast.ArrType{Elem: &ast.AnyType{NamePrv: "Ret"}},
			},
		},
	}
	c3  = &ast.Call{Name: "C3"}
	rc1 = &ast.Call{
		Name: "Rc1",
		Arg:  &ast.AnyType{NamePrv: "Args"},
		Ret:  &ast.StructType{NamePrv: "Rc1Ret"},
	}
	rc2 = &ast.Call{
		Name:  "Rc2",
		Async: true,
		Arg:   &ast.StructType{NamePrv: "Rc2Args"},
	}
	rc3 = &ast.Call{
		Name: "Rc3",
	}
	st1 = &ast.Stream{Name: "S1"}
	st2 = &ast.Stream{
		Name: "S2",
		Arg:  &ast.BaseType{DataType: ast.TypeString},
	}
	st3 = &ast.Stream{
		Name: "S3",
		Ret:  &ast.AnyType{NamePrv: "En1"},
	}
	rst1 = &ast.Stream{
		Name: "Rs1",
		Arg:  &ast.AnyType{NamePrv: "Args"},
		Ret:  &ast.AnyType{NamePrv: "Ret"},
	}
	rst2 = &ast.Stream{
		Name: "Rs2",
		Arg: &ast.MapType{
			Key:   &ast.BaseType{DataType: ast.TypeString},
			Value: &ast.BaseType{DataType: ast.TypeInt},
		},
	}
	rst3 = &ast.Stream{
		Name: "Rs3",
	}
	expSrvc = &ast.Service{
		Calls:   []*ast.Call{c1, c2, c3, rc1, rc2, rc3},
		Streams: []*ast.Stream{st1, st2, st3, rst1, rst2, rst3},
	}

	expTypes = []*ast.Type{
		{
			Name: "Rc1Ret",
			Fields: []*ast.TypeField{
				{Name: "S", DataType: &ast.BaseType{DataType: ast.TypeString}},
				{Name: "I", DataType: &ast.BaseType{DataType: ast.TypeInt}},
				{
					Name: "M",
					DataType: &ast.MapType{
						Key:   &ast.BaseType{DataType: ast.TypeString},
						Value: &ast.BaseType{DataType: ast.TypeInt},
					},
				},
				{Name: "Sl", DataType: &ast.ArrType{Elem: &ast.BaseType{DataType: ast.TypeTime}}},
				{Name: "St", DataType: &ast.AnyType{NamePrv: "Ret"}},
				{
					Name: "Crazy",
					DataType: &ast.MapType{
						Key: &ast.BaseType{DataType: ast.TypeString},
						Value: &ast.ArrType{
							Elem: &ast.ArrType{
								Elem: &ast.MapType{
									Key:   &ast.BaseType{DataType: ast.TypeString},
									Value: &ast.AnyType{NamePrv: "En1"},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "Rc2Args",
			Fields: []*ast.TypeField{
				{Name: "F", DataType: &ast.BaseType{DataType: ast.TypeFloat64}},
				{Name: "B", DataType: &ast.BaseType{DataType: ast.TypeByte}},
				{Name: "U8", DataType: &ast.BaseType{DataType: ast.TypeUInt8}},
				{Name: "U16", DataType: &ast.BaseType{DataType: ast.TypeUInt16}},
				{Name: "U32", DataType: &ast.BaseType{DataType: ast.TypeUInt32}},
				{Name: "U64", DataType: &ast.BaseType{DataType: ast.TypeUInt64}},
			},
		},
		{
			Name: "Args",
			Fields: []*ast.TypeField{
				{Name: "S", DataType: &ast.BaseType{DataType: ast.TypeString}},
				{Name: "I", DataType: &ast.BaseType{DataType: ast.TypeInt}},
				{
					Name: "M",
					DataType: &ast.MapType{
						Key:   &ast.BaseType{DataType: ast.TypeString},
						Value: &ast.BaseType{DataType: ast.TypeInt},
					},
				},
				{Name: "Sl", DataType: &ast.ArrType{Elem: &ast.BaseType{DataType: ast.TypeTime}}},
				{Name: "St", DataType: &ast.AnyType{NamePrv: "Ret"}},
				{
					Name: "Crazy",
					DataType: &ast.MapType{
						Key: &ast.BaseType{DataType: ast.TypeString},
						Value: &ast.ArrType{
							Elem: &ast.ArrType{
								Elem: &ast.MapType{
									Key:   &ast.BaseType{DataType: ast.TypeString},
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
				{Name: "F", DataType: &ast.BaseType{DataType: ast.TypeFloat64}},
				{Name: "B", DataType: &ast.BaseType{DataType: ast.TypeByte}},
				{Name: "U8", DataType: &ast.BaseType{DataType: ast.TypeUInt8}},
				{Name: "U16", DataType: &ast.BaseType{DataType: ast.TypeUInt16}},
				{Name: "U32", DataType: &ast.BaseType{DataType: ast.TypeUInt32}},
				{Name: "U64", DataType: &ast.BaseType{DataType: ast.TypeUInt64}},
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
	t.Parallel()

	p := parse.NewParser()
	tree, err := p.Parse(token.NewReader(strings.NewReader(orbit)))
	require.NoError(t, err)

	// Services.
	require.NotNil(t, tree.Srvc)
	requireEqualService(t, expSrvc, tree.Srvc)

	// Types.
	require.Len(t, tree.Types, len(expTypes))
	for i, expType := range expTypes {
		requireEqualType(t, expType, tree.Types[i])
	}

	// Enums.
	require.Len(t, tree.Enums, len(expEnums))
	for i, expEn := range expEnums {
		requireEqualEnum(t, expEn, tree.Enums[i])
	}

	// Errors.
	require.Len(t, tree.Errs, len(expErrs))
	for i, expErr := range expErrs {
		requireEqualError(t, expErr, tree.Errs[i])
	}
}

//###############//
//### Helpers ###//
//###############//

func requireEqualService(t *testing.T, exp, act *ast.Service) {
	require.Len(t, exp.Calls, len(act.Calls))
	require.Len(t, exp.Streams, len(act.Streams))
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
	requireEqualDataType(t, exp.Arg, act.Arg)
	requireEqualDataType(t, exp.Ret, act.Ret)
}

func requireEqualStream(t *testing.T, exp, act *ast.Stream) {
	require.Exactly(t, exp.Name, act.Name)
	requireEqualDataType(t, exp.Arg, act.Arg)
	requireEqualDataType(t, exp.Ret, act.Ret)
}

func requireEqualType(t *testing.T, exp, act *ast.Type) {
	require.Exactly(t, exp.Name, act.Name)
	require.Len(t, exp.Fields, len(act.Fields))
	for i, exptf := range exp.Fields {
		require.Exactly(t, exptf.Name, act.Fields[i].Name)
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
	case *ast.BaseType, *ast.StructType, *ast.EnumType, *ast.AnyType:
		require.Exactly(t, exp.Decl(), act.Decl())
	case *ast.ArrType:
		requireEqualDataType(t, v.Elem, act.(*ast.ArrType).Elem)
	case *ast.MapType:
		requireEqualDataType(t, v.Key, act.(*ast.MapType).Key)
		requireEqualDataType(t, v.Value, act.(*ast.MapType).Value)
	default:
		t.Fatalf("unknown data type %v", v)
	}
}
