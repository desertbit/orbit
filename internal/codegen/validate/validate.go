package validate

import (
	"github.com/desertbit/orbit/internal/codegen/ast"
)

// Validate validates the given ast.File for various sanity checks and attempts to resolve
// all AnyTypes to either a Struct or Enum Type.
func Validate(f *ast.File) error {
	// Service.
	err := validateService(f)
	if err != nil {
		return err
	}

	// Types.
	for i, t := range f.Types {
		for j := i + 1; j < len(f.Types); j++ {
			// Check for duplicate names.
			if t.Name == f.Types[j].Name {
				return ast.NewErr(t.Line, "type '%s' declared twice", t.Name)
			}
		}

		for j, tf := range t.Fields {
			for k := j + 1; k < len(t.Fields); k++ {
				// Check for duplicate field names.
				if tf.Name == t.Fields[k].Name {
					return ast.NewErr(
						tf.Line, "field '%s' of type '%s' declared twice", tf.Name, t.Name,
					)
				}
			}

			// Resolve all AnyTypes.
			tf.DataType, err = resolveAnyType(tf.DataType, f)
			if err != nil {
				return err
			}
		}
	}

	// Errors.
	for i, e := range f.Errs {
		for j := i + 1; j < len(f.Errs); j++ {
			e2 := f.Errs[j]

			// Check for duplicate names.
			if e.Name == e2.Name {
				return ast.NewErr(e.Line, "error '%s' declared twice", e.Name)
			}

			// Check for valid id.
			if e.ID <= 0 {
				return ast.NewErr(e.Line, "invalid error id, must be greater than 0")
			}

			// Check for duplicate id.
			if e.ID == e2.ID {
				return ast.NewErr(e.Line, "error '%s' has same id as '%s'", e.Name, e2.Name)
			}
		}
	}

	// Enums.
	for i, en := range f.Enums {
		for j := i + 1; j < len(f.Enums); j++ {
			// Check for duplicate name.
			if en.Name == f.Enums[j].Name {
				return ast.NewErr(en.Line, "enum '%s' declared twice", en.Name)
			}
		}
	}

	return nil
}

func resolveAnyType(dt ast.DataType, f *ast.File) (ast.DataType, error) {
	var err error

	switch v := dt.(type) {
	case *ast.AnyType:
		// Resolve the type.
		for _, t := range f.Types {
			if t.Name == v.ID() {
				return &ast.StructType{Name: v.ID(), Pos: v.Pos}, nil
			}
		}
		for _, en := range f.Enums {
			if en.Name == v.ID() {
				return &ast.EnumType{Name: v.ID(), Pos: v.Pos}, nil
			}
		}
		switch v.ID() {
		case ast.TypeByte, ast.TypeString, ast.TypeTime, ast.TypeDuration,
			ast.TypeBool,
			ast.TypeInt, ast.TypeInt8, ast.TypeInt16, ast.TypeInt32, ast.TypeInt64,
			ast.TypeUInt, ast.TypeUInt8, ast.TypeUInt16, ast.TypeUInt32, ast.TypeUInt64,
			ast.TypeFloat32, ast.TypeFloat64:
			return &ast.BaseType{DataType: v.ID(), Pos: v.Pos}, nil
		}

		return nil, ast.NewErr(v.Line, "could not resolve unknown type '%s'", v.ID())
	case *ast.MapType:
		v.Key, err = resolveAnyType(v.Key, f)
		if err != nil {
			return nil, err
		}

		// Key types can only be base or enum types.
		switch v.Key.(type) {
		case *ast.BaseType, *ast.EnumType:
		default:
			return nil, ast.NewErr(v.Line, "invalid map key type '%s'", v.Key.ID())
		}

		v.Value, err = resolveAnyType(v.Value, f)
		if err != nil {
			return nil, err
		}
	case *ast.ArrType:
		v.Elem, err = resolveAnyType(v.Elem, f)
		if err != nil {
			return nil, err
		}
	}

	// No change.
	return dt, nil
}
