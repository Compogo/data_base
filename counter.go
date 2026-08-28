package data_base

import (
	"context"

	dbClient "github.com/Compogo/db-client"
	"github.com/Compogo/db-client/repository"
	"github.com/doug-martin/goqu/v9"
)

// Counter предназначен для подсчёта количества записей в таблице,
// удовлетворяющих переданным фильтрам.
//
// Игнорирует значения NULL в указанной колонке для подсчёта.
// Для nil-фильтров возвращает общее количество записей в таблице.
//
// Пример:
//
//	counter := data_base.NewCounter(db, gen, "users", "id")
//	filters := []*repository.Filter{repository.NewFilter("active", "true", repository.Eq)}
//	count, err := counter.Count(ctx, filters...)
type Counter struct {
	tableName      string
	counterColName string
	db             dbClient.Client
	gen            *goqu.DialectWrapper
}

func NewCounter(db dbClient.Client, gen *goqu.DialectWrapper, tableName string, counterColName string) *Counter {
	return &Counter{tableName: tableName, counterColName: counterColName, db: db, gen: gen}
}

func (c *Counter) Count(ctx context.Context, filters ...*repository.Filter) (uint64, error) {
	queryBuilder := c.gen.Select(goqu.COUNT(goqu.C(c.counterColName))).From(c.tableName).Prepared(true)

	var err error
	var exp goqu.Expression

	for _, filter := range filters {
		exp, err = NewFilter(filter)
		if err != nil {
			return 0, err
		}

		queryBuilder = queryBuilder.Where(exp)
	}

	query, args, err := queryBuilder.ToSQL()
	if err != nil {
		return 0, err
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count uint64
	if err = rows.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
