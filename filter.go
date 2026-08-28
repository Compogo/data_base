package data_base

import (
	"errors"
	"fmt"

	"github.com/Compogo/db-client/repository"
	"github.com/doug-martin/goqu/v9"
)

var FilterTypeUndefined = errors.New("undefined")

// NewFilter преобразует Filter из пакета repository в goqu.Expression,
// который можно использовать для построения SQL-запросов.
// Поддерживает операторы: =, !=, >, >=, <, <=, LIKE, IN.
//
// Возвращает ошибку FilterTypeUndefined, если переданный оператор не поддерживается.
//
// Пример:
//
//	filter := repository.NewFilter("name", "Alice", repository.Eq)
//	expr, err := data_base.NewFilter(filter)
func NewFilter(filter *repository.Filter) (goqu.Expression, error) {
	column := goqu.C(filter.ColumnName)

	switch filter.Comparable {
	case repository.Eq:
		return column.Eq(filter.Value), nil
	case repository.Neq:
		return column.Neq(filter.Value), nil
	case repository.Gt:
		return column.Gt(filter.Value), nil
	case repository.Gte:
		return column.Gte(filter.Value), nil
	case repository.Lt:
		return column.Lt(filter.Value), nil
	case repository.Lte:
		return column.Lte(filter.Value), nil
	case repository.LIKE:
		return column.Like(filter.Value), nil
	case repository.IN:
		return column.In(filter.Value), nil
	}

	return nil, fmt.Errorf("filter type: %s %w", filter.Comparable, FilterTypeUndefined)
}
