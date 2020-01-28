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

package resolve_test

import (
	"strings"
	"testing"

	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/parse"
	"github.com/desertbit/orbit/internal/codegen/resolve"
	"github.com/desertbit/orbit/internal/codegen/token"
	"github.com/stretchr/testify/require"
)

const orbit = `
service s1 {
    call c1 {
        ret: []map[string]Ret
    }

    revcall rc1 {
        args: Args
        ret: {
            s string
            crazy map[string][]En1
        }
    }

    stream s1 {
        ret: En1
    }

    revstream rs1 {
        args: Args
        ret: Ret
    }
}

type Args {
    st Ret
    crazy map[string][]En1
}

type Ret {
    i int
}

enum En1 {
    Val1 = 1
    Val2 = 2
    Val3 = 3
}
`

var (
	c1 = &ast.Call{
		Name: "C1",
		Ret: &ast.ArrType{
			Elem: &ast.MapType{
				Key:   &ast.BaseType{DataType: ast.TypeString},
				Value: &ast.StructType{Name: "Ret"},
			},
		},
	}
	rc1 = &ast.Call{
		Name: "Rc1",
		Rev:  true,
		Args: &ast.StructType{Name: "Args"},
		Ret:  &ast.StructType{Name: "Rc1Ret"},
	}
	st1 = &ast.Stream{
		Name: "S1",
		Ret:  &ast.EnumType{Name: "En1"},
	}
	rst1 = &ast.Stream{
		Name: "Rs1",
		Rev:  true,
		Args: &ast.StructType{Name: "Args"},
		Ret:  &ast.StructType{Name: "Ret"},
	}
	expSrvcs = []*ast.Service{
		{
			Name:    "S1",
			Calls:   []*ast.Call{c1, rc1},
			Streams: []*ast.Stream{st1, rst1},
		},
	}

	expTypes = []*ast.Type{
		{
			Name: "Rc1Ret",
			Fields: []*ast.TypeField{
				{Name: "S", DataType: &ast.BaseType{DataType: ast.TypeString}},
				{
					Name: "Crazy",
					DataType: &ast.MapType{
						Key: &ast.BaseType{DataType: ast.TypeString},
						Value: &ast.ArrType{
							Elem: &ast.EnumType{Name: "En1"},
						},
					},
				},
			},
		},
		{
			Name: "Args",
			Fields: []*ast.TypeField{
				{Name: "St", DataType: &ast.StructType{Name: "Ret"}},
				{
					Name: "Crazy",
					DataType: &ast.MapType{
						Key: &ast.BaseType{DataType: ast.TypeString},
						Value: &ast.ArrType{
							Elem: &ast.EnumType{Name: "En1"},
						},
					},
				},
			},
		},
		{
			Name: "Ret",
			Fields: []*ast.TypeField{
				{Name: "I", DataType: &ast.BaseType{DataType: ast.TypeInt}},
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
)

func TestResolve(t *testing.T) {
	t.Parallel()

	p := parse.NewParser()
	srvcs, types, errs, enums, err := p.Parse(token.NewReader(strings.NewReader(orbit)))
	require.NoError(t, err)

	err = resolve.Resolve(srvcs, types, errs, enums)
	require.NoError(t, err)
	// Services.
	require.Len(t, srvcs, len(expSrvcs))
	for i, expSrvc := range expSrvcs {
		requireEqualService(t, expSrvc, srvcs[i])
	}

	// Types.
	require.Len(t, types, len(expTypes))
	for i, expType := range expTypes {
		requireEqualType(t, expType, types[i])
	}
}

//###############//
//### Helpers ###//
//###############//

func requireEqualService(t *testing.T, exp, act *ast.Service) {
	require.Equal(t, exp.Name, act.Name)
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
	require.Exactly(t, exp.Rev, act.Rev)
	requireEqualDataType(t, exp.Args, act.Args)
	requireEqualDataType(t, exp.Ret, act.Ret)
}

func requireEqualStream(t *testing.T, exp, act *ast.Stream) {
	require.Exactly(t, exp.Name, act.Name)
	require.Exactly(t, exp.Rev, act.Rev)
	requireEqualDataType(t, exp.Args, act.Args)
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

func requireEqualDataType(t *testing.T, exp, act ast.DataType) {
	if exp == nil && act == nil {
		return
	}

	require.IsType(t, exp, act)
	switch v := exp.(type) {
	case *ast.BaseType, *ast.StructType, *ast.EnumType, *ast.AnyType:
		require.Exactly(t, exp.String(), act.String())
	case *ast.ArrType:
		requireEqualDataType(t, v.Elem, act.(*ast.ArrType).Elem)
	case *ast.MapType:
		requireEqualDataType(t, v.Key, act.(*ast.MapType).Key)
		requireEqualDataType(t, v.Value, act.(*ast.MapType).Value)
	default:
		t.Fatalf("unknown data type %v", v)
	}
}
