package dbtype

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type TextArray[T ~string] []T

func (a TextArray[T]) Value() (driver.Value, error) {
	if a == nil {
		return pgtype.FlatArray[string]{}, nil
	}

	arr := make(pgtype.FlatArray[string], len(a))
	for i, v := range a {
		arr[i] = string(v)
	}

	return arr, nil
}

func (a *TextArray[T]) Scan(src any) error {
	if src == nil {
		*a = nil
		return nil
	}

	var buf []byte
	switch v := src.(type) {
	case []byte:
		buf = v
	case string:
		buf = []byte(v)
	default:
		return fmt.Errorf("unsupported type %T", src)
	}

	m := pgtype.NewMap()

	var elements []string
	t, ok := m.TypeForValue(&elements)
	if !ok {
		return fmt.Errorf("cannot find pg type for []string")
	}

	if err := m.Scan(t.OID, pgtype.BinaryFormatCode, buf, &elements); err != nil {
		return err
	}

	res := make(TextArray[T], len(elements))
	for i, e := range elements {
		res[i] = T(e)
	}

	*a = res
	return nil
}
