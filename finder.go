package data_base

import (
	"context"

	dbClient "github.com/Compogo/db-client"
	"github.com/Compogo/db-client/repository"
	"github.com/doug-martin/goqu/v9"
)

// Finder предназначен для поиска записей в базе данных с поддержкой фильтрации
// и сортировки. Возвращает срез указателей на модели типа T.
//
// Нулевое значение структуры использовать нельзя. Для создания экземпляра
// используйте функцию NewFinder, передавая в неё экземпляр клиента БД,
// генератор SQL, функцию для маппинга строки результата в модель и название таблицы.
//
// Пример:
//
//	rowMapper := func(rows *sql.Rows) (*User, error) {
//	    var u User
//	    err := rows.Scan(&u.ID, &u.Name, &u.Email)
//	    return &u, err
//	}
//
//	finder := data_base.NewFinder[User](db, goqu.Dialect("postgres"), rowMapper, "users", "id", "name", "email")
//	filters := []*repository.Filter{repository.NewFilter("name", "Alice", repository.Eq)}
//	users, err := finder.Find(ctx, nil, filters...)
type Finder[T any] struct {
	tableName string
	columns   []any
	scanner   Scanner[T]
	db        dbClient.Client
	gen       *goqu.DialectWrapper
}

func NewFinder[T any](db dbClient.Client, gen *goqu.DialectWrapper, scanner Scanner[T], tableName string, columns ...any) *Finder[T] {
	return &Finder[T]{tableName: tableName, columns: columns, db: db, gen: gen, scanner: scanner}
}

func (f *Finder[T]) Find(ctx context.Context, sorts []*repository.Sort, filters ...*repository.Filter) ([]*T, error) {
	selectDataset := f.gen.Select(f.columns...).From(f.tableName).Prepared(true)

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

	query, args, err := selectDataset.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := f.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*T
	for rows.Next() {
		model, err := f.scanner(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}
