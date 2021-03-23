package validate_test

import (
	"testing"
	"time"

	"github.com/desertbit/orbit/internal/codegen/ast"
	"github.com/desertbit/orbit/internal/codegen/lexer"
	"github.com/desertbit/orbit/internal/codegen/validate"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Run("valid", testValidateValid)
}

func testValidateValid(t *testing.T) {
	t.Parallel()

	// Copied from parser.parser_test.go
	var (
		c2Timeout          = 500 * time.Millisecond
		c2MaxArgSize int64 = 154 * 1024
		c2MaxRetSize int64 = 5 * 1024 * 1024
	)
	var (
		expVersion = &ast.Version{Value: 1, Pos: lexer.Pos{Line: 1, Column: 9}}
		c1         = &ast.Call{
			Name: "C1",
			Arg:  &ast.StructType{Name: "C1Arg"},
			Ret:  &ast.StructType{Name: "C1Ret"},
		}
		c2 = &ast.Call{
			Name:       "C2",
			Async:      true,
			Arg:        &ast.StructType{Name: "C2Arg"},
			Ret:        &ast.StructType{Name: "C2Ret"},
			Timeout:    &c2Timeout,
			MaxArgSize: &c2MaxArgSize,
			MaxRetSize: &c2MaxRetSize,
		}
		c3  = &ast.Call{Name: "C3"}
		rc1 = &ast.Call{
			Name: "Rc1",
			Arg:  &ast.AnyType{Name: "Arg"},
			Ret:  &ast.StructType{Name: "Rc1Ret"},
		}
		rc2 = &ast.Call{
			Name:  "Rc2",
			Async: true,
			Arg:   &ast.StructType{Name: "Rc2Arg"},
		}
		rc3 = &ast.Call{
			Name: "Rc3",
		}
		st1 = &ast.Stream{Name: "S1"}
		st2 = &ast.Stream{
			Name: "S2",
			Arg:  &ast.StructType{Name: "S2Arg"},
		}
		st3 = &ast.Stream{
			Name: "S3",
			Ret:  &ast.AnyType{Name: "Ret"},
		}
		rst1 = &ast.Stream{
			Name: "Rs1",
			Arg:  &ast.AnyType{Name: "Arg"},
			Ret:  &ast.AnyType{Name: "Ret"},
		}
		rst2 = &ast.Stream{
			Name: "Rs2",
		}
		expSrvc = &ast.Service{
			Calls:   []*ast.Call{c1, c2, c3, rc1, rc2, rc3},
			Streams: []*ast.Stream{st1, st2, st3, rst1, rst2},
		}

		c1Arg = &ast.Type{
			Name: "C1Arg",
			Fields: []*ast.TypeField{
				{Name: "Id", DataType: &ast.AnyType{Name: "int"}, StructTag: "json:\"ID\" yaml:\"id\""},
			},
		}
		c1Ret = &ast.Type{
			Name: "C1Ret",
			Fields: []*ast.TypeField{
				{Name: "Sum", DataType: &ast.AnyType{Name: "float32"}},
			},
		}
		c2Arg = &ast.Type{
			Name: "C2Arg",
			Fields: []*ast.TypeField{
				{Name: "Ts", DataType: &ast.AnyType{Name: "time"}},
			},
		}
		c2Ret = &ast.Type{
			Name: "C2Ret",
			Fields: []*ast.TypeField{
				{
					Name: "Data",
					// []map[string][]Ret
					DataType: &ast.ArrType{
						Elem: &ast.MapType{
							Key:   &ast.AnyType{Name: "string"},
							Value: &ast.ArrType{Elem: &ast.AnyType{Name: "Ret"}},
						},
					},
				},
			},
		}
		rc1Ret = &ast.Type{
			Name: "Rc1Ret",
			Fields: []*ast.TypeField{
				{Name: "S", DataType: &ast.AnyType{Name: "string"}},
				{Name: "I", DataType: &ast.AnyType{Name: "int"}},
				{
					Name: "M",
					DataType: &ast.MapType{
						Key:   &ast.AnyType{Name: "string"},
						Value: &ast.AnyType{Name: "int"},
					},
				},
				{Name: "Sl", DataType: &ast.ArrType{Elem: &ast.AnyType{Name: "time"}}},
				{Name: "St", DataType: &ast.AnyType{Name: "Ret"}},
				{
					Name: "Crazy",
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
		}
		rc2Arg = &ast.Type{
			Name: "Rc2Arg",
			Fields: []*ast.TypeField{
				{Name: "F", DataType: &ast.AnyType{Name: "float64"}},
				{Name: "B", DataType: &ast.AnyType{Name: "byte"}},
				{Name: "U8", DataType: &ast.AnyType{Name: "uint8"}},
				{Name: "U16", DataType: &ast.AnyType{Name: "uint16"}},
				{Name: "U32", DataType: &ast.AnyType{Name: "uint32"}},
				{Name: "U64", DataType: &ast.AnyType{Name: "uint64"}},
			},
		}
		s2Arg = &ast.Type{
			Name: "S2Arg",
			Fields: []*ast.TypeField{
				{Name: "Id", DataType: &ast.AnyType{Name: "string"}, StructTag: "validator:\"required\""},
			},
		}
		arg = &ast.Type{
			Name: "Arg",
			Fields: []*ast.TypeField{
				{Name: "S", DataType: &ast.AnyType{Name: "string"}, StructTag: "json:\"STRING\""},
				{Name: "I", DataType: &ast.AnyType{Name: "int"}},
				{
					Name: "M",
					DataType: &ast.MapType{
						Key:   &ast.AnyType{Name: "string"},
						Value: &ast.AnyType{Name: "int"},
					},
				},
				{Name: "Sl", DataType: &ast.ArrType{Elem: &ast.AnyType{Name: "time"}}},
				{Name: "Dur", DataType: &ast.AnyType{Name: "duration"}},
				{Name: "St", DataType: &ast.AnyType{Name: "Ret"}},
				{
					Name: "Crazy",
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
		}
		ret = &ast.Type{
			Name: "Ret",
			Fields: []*ast.TypeField{
				{Name: "F", DataType: &ast.AnyType{Name: "float64"}},
				{Name: "B", DataType: &ast.AnyType{Name: "byte"}},
				{Name: "U8", DataType: &ast.AnyType{Name: "uint8"}},
				{Name: "U16", DataType: &ast.AnyType{Name: "uint16"}},
				{Name: "U32", DataType: &ast.AnyType{Name: "uint32"}},
				{Name: "U64", DataType: &ast.AnyType{Name: "uint64"}},
			},
		}
		expTypes = []*ast.Type{c1Arg, c1Ret, c2Arg, c2Ret, rc1Ret, rc2Arg, s2Arg, arg, ret}

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

	// Call validate.
	f := &ast.File{Version: expVersion, Srvc: expSrvc, Types: expTypes, Enums: expEnums, Errs: expErrs}
	require.NoError(t, validate.Validate(f))

	// Check, if all any types have been resolved.
	require.IsType(t, &ast.StructType{}, rc1.Arg)
	require.IsType(t, &ast.StructType{}, st3.Ret)
	require.IsType(t, &ast.StructType{}, rst1.Arg)
	require.IsType(t, &ast.StructType{}, rst1.Ret)
	require.IsType(t, &ast.BaseType{}, c1Arg.Fields[0].DataType)
	require.IsType(t, &ast.BaseType{}, c1Ret.Fields[0].DataType)
	require.IsType(t, &ast.BaseType{}, c2Arg.Fields[0].DataType)
	require.IsType(t, &ast.BaseType{}, c2Ret.Fields[0].DataType.(*ast.ArrType).Elem.(*ast.MapType).Key)
	require.IsType(t, &ast.BaseType{}, rc1Ret.Fields[0].DataType)
	require.IsType(t, &ast.BaseType{}, rc1Ret.Fields[1].DataType)
	require.IsType(t, &ast.BaseType{}, rc1Ret.Fields[2].DataType.(*ast.MapType).Key)
	require.IsType(t, &ast.BaseType{}, rc1Ret.Fields[2].DataType.(*ast.MapType).Value)
	require.IsType(t, &ast.BaseType{}, rc1Ret.Fields[3].DataType.(*ast.ArrType).Elem)
	require.IsType(t, &ast.StructType{}, rc1Ret.Fields[4].DataType)
	require.IsType(t, &ast.BaseType{}, rc1Ret.Fields[5].DataType.(*ast.MapType).Key)
	require.IsType(t, &ast.BaseType{}, rc1Ret.Fields[5].DataType.(*ast.MapType).Value.(*ast.ArrType).Elem.(*ast.ArrType).Elem.(*ast.MapType).Key)
	require.IsType(t, &ast.EnumType{}, rc1Ret.Fields[5].DataType.(*ast.MapType).Value.(*ast.ArrType).Elem.(*ast.ArrType).Elem.(*ast.MapType).Value)
}
