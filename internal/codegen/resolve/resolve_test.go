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
version 5

service {
    call c1 {
        ret: []map[string]Ret
    }

    call rc1 {
        arg: Args
        ret: {
            s string
            crazy map[string][]En1
        }
    }

    stream s1 {
        ret: En1
    }

    stream rs1 {
        arg: Args
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
	version = 5
	c1      = &ast.Call{
		Name: "C1",
		Ret: &ast.ArrType{
			Elem: &ast.MapType{
				Key:   &ast.BaseType{DataType: ast.TypeString},
				Value: &ast.StructType{NamePrv: "Ret"},
			},
		},
	}
	rc1 = &ast.Call{
		Name: "Rc1",
		Arg:  &ast.StructType{NamePrv: "Args"},
		Ret:  &ast.StructType{NamePrv: "Rc1Ret"},
	}
	st1 = &ast.Stream{
		Name: "S1",
		Ret:  &ast.EnumType{NamePrv: "En1"},
	}
	rst1 = &ast.Stream{
		Name: "Rs1",
		Arg:  &ast.StructType{NamePrv: "Args"},
		Ret:  &ast.StructType{NamePrv: "Ret"},
	}
	expSrvc = &ast.Service{
		Calls:   []*ast.Call{c1, rc1},
		Streams: []*ast.Stream{st1, rst1},
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
							Elem: &ast.EnumType{NamePrv: "En1"},
						},
					},
				},
			},
		},
		{
			Name: "Args",
			Fields: []*ast.TypeField{
				{Name: "St", DataType: &ast.StructType{NamePrv: "Ret"}},
				{
					Name: "Crazy",
					DataType: &ast.MapType{
						Key: &ast.BaseType{DataType: ast.TypeString},
						Value: &ast.ArrType{
							Elem: &ast.EnumType{NamePrv: "En1"},
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
	tree, err := p.Parse(token.NewReader(strings.NewReader(orbit)))
	require.NoError(t, err)

	err = resolve.Resolve(tree)
	require.NoError(t, err)

	// Version.
	require.Exactly(t, version, tree.Version)

	// Service.
	require.NotNil(t, tree.Srvc)
	requireEqualService(t, expSrvc, tree.Srvc)

	// Types.
	require.Len(t, tree.Types, len(expTypes))
	for i, expType := range expTypes {
		requireEqualType(t, expType, tree.Types[i])
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
