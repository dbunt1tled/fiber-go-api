package storage

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/dbunt1tled/fiber-go/pkg/f"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type Scanner[T Model] func(pgx.Row) (T, error)

type RowsScanner[T Model] func(pgx.Rows) ([]T, error)

type RecordBuilder[T Model] func(T) goqu.Record

type Repository[T Model] struct {
	db            *pgxpool.Pool
	dialect       goqu.DialectWrapper
	table         string
	scanner       Scanner[T]
	rowsScanner   RowsScanner[T]
	recordBuilder RecordBuilder[T]
}

func NewRepository[T Model](
	pool *pgxpool.Pool,
	table string,
	scanner Scanner[T],
	rowsScanner RowsScanner[T],
	recordBuilder RecordBuilder[T],
) *Repository[T] {
	return &Repository[T]{
		db:            pool,
		dialect:       goqu.Dialect("postgres"),
		table:         table,
		scanner:       scanner,
		rowsScanner:   rowsScanner,
		recordBuilder: recordBuilder,
	}
}

func (r *Repository[T]) FindByID(ctx context.Context, id uuid.UUID) (T, error) {
	return r.findByIDWithQuerier(ctx, r.db, id)
}

func (r *Repository[T]) List(ctx context.Context, opts ...QueryOption) ([]T, error) {
	cfg := &queryConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	filter := r.buildFilter(cfg)
	return r.findWithQuerier(ctx, r.db, filter)
}

func (r *Repository[T]) One(ctx context.Context, opts ...QueryOption) (T, error) {
	var zero T
	cfg := &queryConfig{limit: 1}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.limit = 1
	cfg.offset = 0

	filter := r.buildFilter(cfg)

	entities, err := r.findWithQuerier(ctx, r.db, filter)
	if err != nil {
		return zero, err
	}

	if len(entities) == 0 {
		return zero, nil
	}

	return entities[0], nil
}

func (r *Repository[T]) Paginate(
	ctx context.Context,
	page int,
	perPage int,
	opts ...QueryOption,
) (*Paginator[T], error) {

	page, perPage = NormalizePagination(page, perPage)

	cfg := &queryConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var (
		items      []T
		totalCount int64
		totalPages int
	)

	g, cx := errgroup.WithContext(ctx)

	g.Go(func() error {
		cfg.limit = perPage
		cfg.offset = (page - 1) * perPage

		filter := r.buildFilter(cfg)
		result, err := r.findWithQuerier(cx, r.db, filter)
		if err != nil {
			return fmt.Errorf("fetch items: %w", err)
		}
		items = result
		return nil
	})

	g.Go(func() error {
		result, err := r.countWithConfig(cx, cfg)
		if err != nil {
			return fmt.Errorf("count total: %w", err)
		}
		totalCount = result
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	totalPages = int(totalCount) / perPage
	if int(totalCount)%perPage != 0 {
		totalPages++
	}

	return &Paginator[T]{
		Items:      items,
		Total:      totalCount,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}, nil
}

func (r *Repository[T]) Insert(ctx context.Context, entity T) (T, error) {
	var zero T
	err := r.insertWithQuerier(ctx, r.db, entity)
	if err != nil {
		return zero, fmt.Errorf("insert entity: %w", err)
	}

	return r.findByIDWithQuerier(ctx, r.db, entity.GetID())
}

func (r *Repository[T]) InsertBatch(ctx context.Context, entities []T) error {
	return r.insertBatchWithQuerier(ctx, r.db, entities)
}

func (r *Repository[T]) Update(ctx context.Context, entity T) (T, error) {
	var zero T
	err := r.updateWithQuerier(ctx, r.db, entity)
	if err != nil {
		return zero, err
	}
	return r.FindByID(ctx, entity.GetID())
}

func (r *Repository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	return r.deleteWithQuerier(ctx, r.db, id)
}

func (r *Repository[T]) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	return r.deleteBatchWithQuerier(ctx, r.db, ids)
}

func (r *Repository[T]) Count(ctx context.Context, opts ...QueryOption) (int64, error) {
	cfg := &queryConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	return r.countWithConfig(ctx, cfg)
}

func (r *Repository[T]) countWithConfig(ctx context.Context, cfg *queryConfig) (int64, error) {
	query := r.dialect.From(r.table).Select(goqu.COUNT("*"))

	for _, rule := range cfg.rules {
		query = r.applyRule(query, rule)
	}

	sql, args, err := query.ToSQL()
	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}

	var count int64
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("scan count: %w", err)
	}

	return count, nil
}

func (r *Repository[T]) findByIDWithQuerier(ctx context.Context, q Querier, id uuid.UUID) (T, error) {
	var zero T

	sql, args, err := r.dialect.
		From(r.table).
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return zero, fmt.Errorf("build query: %w", err)
	}

	entity, err := r.scanner(q.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return zero, nil
		}
		return zero, fmt.Errorf("scan row: %w", err)
	}

	return entity, nil
}

func (r *Repository[T]) findWithQuerier(ctx context.Context, q Querier, filter *Filter) ([]T, error) {
	query := r.applyFilter(r.dialect.From(r.table), filter)

	sql, args, err := query.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}
	fmt.Println(sql, args)
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	entities, err := r.rowsScanner(rows)
	if err != nil {
		return nil, fmt.Errorf("scan rows: %w", err)
	}

	return entities, nil
}

func (r *Repository[T]) insertWithQuerier(ctx context.Context, q Querier, entity T) error {
	record := r.recordBuilder(entity)
	record["id"] = entity.GetID()

	sql, args, err := r.dialect.
		Insert(r.table).
		Rows(record).
		ToSQL()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	if _, err := q.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	return nil
}

func (r *Repository[T]) insertBatchWithQuerier(ctx context.Context, q Querier, entities []T) error {
	if len(entities) == 0 {
		return nil
	}

	records := make([]interface{}, 0, len(entities))
	for i := range entities {
		if entities[i].GetID() == uuid.Nil {
			entities[i].SetID(uuid.New())
		}

		record := r.recordBuilder(entities[i])
		record["id"] = entities[i].GetID()
		records = append(records, record)
	}

	sql, args, err := r.dialect.
		Insert(r.table).
		Rows(records...).
		ToSQL()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	if _, err := q.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	return nil
}

func (r *Repository[T]) updateWithQuerier(ctx context.Context, q Querier, entity T) error {
	record := r.recordBuilder(entity)

	sql, args, err := r.dialect.
		Update(r.table).
		Set(record).
		Where(goqu.Ex{"id": entity.GetID()}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	result, err := q.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("entity not found")
	}

	return nil
}

func (r *Repository[T]) deleteWithQuerier(ctx context.Context, q Querier, id uuid.UUID) error {
	sql, args, err := r.dialect.
		Delete(r.table).
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	result, err := q.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("entity not found")
	}

	return nil
}

func (r *Repository[T]) deleteBatchWithQuerier(ctx context.Context, q Querier, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	sql, args, err := r.dialect.
		Delete(r.table).
		Where(goqu.Ex{"id": ids}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	if _, err := q.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	return nil
}

func (r *Repository[T]) buildFilter(cfg *queryConfig) *Filter {
	if cfg == nil {
		return nil
	}

	return &Filter{
		Rules:   cfg.rules,
		OrderBy: cfg.orderBy,
		Limit:   cfg.limit,
		Offset:  cfg.offset,
	}
}

func (r *Repository[T]) applyFilter(query *goqu.SelectDataset, filter *Filter) *goqu.SelectDataset {
	if filter == nil {
		return query
	}

	for _, rule := range filter.Rules {
		query = r.applyRule(query, rule)
	}

	for _, order := range filter.OrderBy {
		if order.Descending {
			query = query.Order(goqu.C(order.Field).Desc())
		} else {
			query = query.Order(goqu.C(order.Field).Asc())
		}
	}

	if filter.Limit > 0 {
		query = query.Limit(uint(filter.Limit))
	}

	if filter.Offset > 0 {
		query = query.Offset(uint(filter.Offset))
	}

	return query
}

func (r *Repository[T]) applyRule(query *goqu.SelectDataset, rule Rule) *goqu.SelectDataset {
	if f.IsNil(rule.Value) && (rule.Operation != OpIsNull && rule.Operation != OpIsNotNull) {
		return query
	}

	col := goqu.C(rule.Field)

	switch rule.Operation {
	case OpEqual:
		return query.Where(goqu.Ex{rule.Field: rule.Value})
	case OpNotEqual:
		return query.Where(col.Neq(rule.Value))
	case OpGreaterThan:
		return query.Where(col.Gt(rule.Value))
	case OpGreaterThanOrEqual:
		return query.Where(col.Gte(rule.Value))
	case OpLessThan:
		return query.Where(col.Lt(rule.Value))
	case OpLessThanOrEqual:
		return query.Where(col.Lte(rule.Value))
	case OpLike:
		return query.Where(col.Like(rule.Value))
	case OpILike:
		return query.Where(col.ILike(rule.Value))
	case OpIn:
		return query.Where(goqu.Ex{rule.Field: rule.Value})
	case OpNotIn:
		return query.Where(col.NotIn(rule.Value))
	case OpIsNull:
		return query.Where(col.IsNull())
	case OpIsNotNull:
		return query.Where(col.IsNotNull())
	case OpContains:
		// Array contains: column @> value
		arrayValue := convertToArrayExpression(rule.Value)
		return query.Where(goqu.L("? @> ?", col, arrayValue))
	case OpContainedBy:
		// Array contained by: column <@ value
		arrayValue := convertToArrayExpression(rule.Value)
		return query.Where(goqu.L("? <@ ?", col, arrayValue))
	case OpOverlaps:
		// Array overlaps: column && value
		arrayValue := convertToArrayExpression(rule.Value)
		return query.Where(goqu.L("? && ?", col, arrayValue))
	case OpJsonContains:
		// JSONB contains: column @> value
		// For JSONB, pass the value as-is (it should be json.RawMessage or string)
		return query.Where(goqu.L("? @> ?::jsonb", col, rule.Value))
	case OpJsonExists:
		// JSONB key exists: column ? value
		return query.Where(goqu.L("? ?? ?", col, rule.Value))
	default:
		return query
	}
}

func convertToArrayExpression(value interface{}) goqu.Expression {
	if value == nil {
		return goqu.L("NULL")
	}

	rv := reflect.ValueOf(value)

	// Handle already converted array expression
	if expr, ok := value.(goqu.Expression); ok {
		return expr
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		// Convert slice to properly typed array
		length := rv.Len()
		if length == 0 {
			// Empty array - need to infer type
			return goqu.L("ARRAY[]::text[]")
		}

		// Use ToPgArray for proper array formatting
		switch v := value.(type) {
		case []string:
			return ToPgArray(v)
		case []int:
			return ToPgArray(v)
		case []int64:
			return ToPgArray(v)
		case []uuid.UUID:
			return ToPgArray(v)
		case []interface{}:
			// Try to convert []interface{} to typed slice
			if length > 0 {
				first := rv.Index(0).Interface()
				switch first.(type) {
				case string:
					strSlice := make([]string, length)
					for i := 0; i < length; i++ {
						strSlice[i] = rv.Index(i).Interface().(string)
					}
					return ToPgArray(strSlice)
				case int:
					intSlice := make([]int, length)
					for i := 0; i < length; i++ {
						intSlice[i] = rv.Index(i).Interface().(int)
					}
					return ToPgArray(intSlice)
				case int64:
					int64Slice := make([]int64, length)
					for i := 0; i < length; i++ {
						int64Slice[i] = rv.Index(i).Interface().(int64)
					}
					return ToPgArray(int64Slice)
				case uuid.UUID:
					uuidSlice := make([]uuid.UUID, length)
					for i := 0; i < length; i++ {
						uuidSlice[i] = rv.Index(i).Interface().(uuid.UUID)
					}
					return ToPgArray(uuidSlice)
				}
			}
		}

		// Fallback: create array literal manually
		values := make([]interface{}, length)
		for i := 0; i < length; i++ {
			values[i] = rv.Index(i).Interface()
		}

		placeholders := strings.Repeat("?,", length)
		if len(placeholders) > 0 {
			placeholders = placeholders[:len(placeholders)-1]
		}

		return goqu.L(fmt.Sprintf("ARRAY[%s]", placeholders), values...)

	default:
		// Single value - wrap in array
		return goqu.L("ARRAY[?]", value)
	}
}


func ScanRowsWithScanner[T any](rows pgx.Rows, scanner func(pgx.Row) (T, error)) ([]T, error) {
	var results []T

	for rows.Next() {
		result, err := scanner(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func ToPgArray[T any](slice []T) goqu.Expression {
	if len(slice) == 0 {
		var zero T
		pgType := inferPgType(zero)
		return goqu.L(fmt.Sprintf("ARRAY[]::%s[]", pgType))
	}

	vals := make([]any, len(slice))
	pgType := inferPgType(slice[0])

	for i, v := range slice {
		rv := reflect.ValueOf(v)

		if _, ok := any(v).(uuid.UUID); ok {
			vals[i] = v
			continue
		}

		switch rv.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.Bool:
			vals[i] = v

		case reflect.Array:
			if rv.Type() == reflect.TypeOf(uuid.UUID{}) {
				vals[i] = v
			} else {
				panic(fmt.Sprintf("ToPgArray: unsupported array type %T", v))
			}

		default:
			if rv.CanConvert(reflect.TypeOf("")) {
				vals[i] = rv.Convert(reflect.TypeOf("")).String()
			} else {
				panic(fmt.Sprintf("ToPgArray: unsupported element type %T", v))
			}
		}
	}

	placeholders := strings.Repeat("?,", len(vals))
	placeholders = placeholders[:len(placeholders)-1]

	arrayLiteral := fmt.Sprintf("ARRAY[%s]::%s[]", placeholders, pgType)

	return goqu.L(arrayLiteral, vals...)
}

func inferPgType[T any](v T) string {
	if _, ok := any(v).(uuid.UUID); ok {
		return "uuid"
	}

	rv := reflect.ValueOf(v)

	if rv.Type() == reflect.TypeOf(uuid.UUID{}) {
		return "uuid"
	}

	switch rv.Kind() {
	case reflect.String:
		return "text"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return "integer"
	case reflect.Int64:
		return "bigint"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "integer"
	case reflect.Uint64:
		return "bigint"
	case reflect.Float32:
		return "real"
	case reflect.Float64:
		return "double precision"
	case reflect.Bool:
		return "boolean"
	case reflect.Array:
		if rv.Type() == reflect.TypeOf(uuid.UUID{}) {
			return "uuid"
		}
		return "text"
	default:
		if rv.CanConvert(reflect.TypeOf("")) {
			return "text"
		}
		return "text"
	}
}
