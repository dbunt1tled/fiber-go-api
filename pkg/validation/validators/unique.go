package validators

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UniqueFieldValidator struct {
	db *pgxpool.Pool
}

func NewUniqueFieldValidator(db *pgxpool.Pool) *UniqueFieldValidator {
	return &UniqueFieldValidator{db: db}
}

func (uv *UniqueFieldValidator) Validate(fl validator.FieldLevel) bool {
	// Format: unique_db=table.column[.exclude_id]
	params := strings.Split(fl.Param(), ".")
	if len(params) < 2 {
		return false
	}

	tableName := params[0]
	columnName := params[1]

	// Prepare the query
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", tableName, columnName)
	args := []interface{}{fl.Field().Interface()}

	if len(params) > 2 && params[2] == "exclude_id" {
		currentStruct := fl.Parent()
		if currentStruct.Kind() == reflect.Ptr {
			currentStruct = currentStruct.Elem()
		}

		// Try to get the ID field
		idField := currentStruct.FieldByName("ID")
		if !idField.IsValid() || idField.IsZero() {
			return true // If no ID field, can't exclude anything
		}

		query += " AND id != $2"
		args = append(args, idField.Interface())
	}

	var count int
	err := uv.db.QueryRow(context.Background(), query, args...).Scan(&count)
	if err != nil {
		// In case of error, fail closed (assume not unique)
		return false
	}

	return count == 0
}
