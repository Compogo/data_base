package data_base

import (
	"github.com/Compogo/db-client/repository"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// NewSort преобразует Sort из пакета repository в goqu.OrderedExpression
// для сортировки результатов запроса.
//
// Поддерживает направления: ASC (по умолчанию) и DESC.
//
// Пример:
//
//	sort := &repository.Sort{ColumnName: "created_at", Direction: repository.DESC}
//	expr := data_base.NewSort(sort)
func NewSort(sort *repository.Sort) exp.OrderedExpression {
	column := goqu.C(sort.ColumnName)

	if sort.Direction == repository.DESC {
		return column.Desc()
	}

	return column.Asc()
}
