package data_base

import (
	"context"

	dbClient "github.com/Compogo/db-client"
	"github.com/Compogo/db-client/repository"
	"github.com/doug-martin/goqu/v9"
)

// Deleter предназначен для удаления записей из таблицы по переданным фильтрам.
// Для выполнения физического удаления (DELETE) используйте фильтры.
// Для мягкого удаления рекомендуется использовать Updater.
//
// Возвращает ошибку, если удаление не удалось выполнить.
//
// Пример:
//
//	deleter := data_base.NewDeleter[*User](db, gen, "users")
//	filters := []*repository.Filter{repository.NewFilter("id", 123, repository.Eq)}
//	err := deleter.Delete(ctx, filters...)
type Deleter[T any] struct {
	tableName string
	db        dbClient.Client
	gen       *goqu.DialectWrapper
}

func NewDeleter[T any](db dbClient.Client, gen *goqu.DialectWrapper, tableName string) *Deleter[T] {
	return &Deleter[T]{tableName: tableName, db: db, gen: gen}
}

func (d *Deleter[T]) Delete(ctx context.Context, filters ...*repository.Filter) error {
	deleteDataset := d.gen.Delete(d.tableName).Prepared(true)

	var err error
	var exp goqu.Expression
	for _, filter := range filters {
		exp, err = NewFilter(filter)
		if err != nil {
			return err
		}

		deleteDataset = deleteDataset.Where(exp)
	}

	query, args, err := deleteDataset.ToSQL()
	if err != nil {
		return err
	}

	_, err = d.db.ExecContext(ctx, query, args...)
	return err
}
