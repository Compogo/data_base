package data_base

import (
	"context"

	dbClient "github.com/Compogo/db-client"
	"github.com/Compogo/db-client/repository"
	"github.com/doug-martin/goqu/v9"
)

// Pager предназначен для постраничного получения данных с фильтрацией и сортировкой.
// Является расширением Finder и использует тот же механизм маппинга строк в модели.
//
// Поддерживает все виды фильтров и сортировок из пакета repository.
// Для работы с пагинацией используйте структуру repository.Page.
//
// Пример:
//
//	pager := data_base.NewPager(db, gen, rowMapper, "users", "id", "name")
//	page := &repository.Page{Number: 0, Limit: 10}
//	users, err := pager.Page(ctx, page, sorts, filters...)
type Pager[T any] struct {
	tableName  string
	columns    []any
	rowToModel RowToModelFunc[T]
	db         dbClient.Client
	gen        *goqu.DialectWrapper
}

func NewPager[T any](db dbClient.Client, gen *goqu.DialectWrapper, rowToModel RowToModelFunc[T], tableName string, columns ...any) *Pager[T] {
	return &Pager[T]{tableName: tableName, columns: columns, rowToModel: rowToModel, db: db, gen: gen}
}

func (p *Pager[T]) Page(ctx context.Context, page *repository.Page, sorts []*repository.Sort, filters ...*repository.Filter) ([]*T, error) {
	selectDataset := p.gen.Select(p.columns...).From(p.tableName).Prepared(true)

	var err error
	var exp goqu.Expression
	for _, filter := range filters {
		exp, err = NewFilter(filter)
		if err != nil {
			return nil, err
		}

		selectDataset = selectDataset.Where(exp)
	}

	for _, sort := range sorts {
		selectDataset = selectDataset.OrderAppend(NewSort(sort))
	}

	query, args, err := selectDataset.Limit(uint(page.Limit)).Offset(uint(page.Limit * page.Number)).ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*T
	for rows.Next() {
		model, err := p.rowToModel(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}
