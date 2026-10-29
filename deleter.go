package data_base

import (
	"context"
	"database/sql"

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
//	deleter := data_base.NewDeleter(db, gen, "users")
//	filters := []*repository.Filter{repository.NewFilter("id", 123, repository.Eq)}
//	err := deleter.Delete(ctx, filters...)
type Deleter struct {
	tableName string
	db        dbClient.Client
	gen       *goqu.DialectWrapper

	preHooks  HookManager[repository.Filter]
	postHooks HookManager[repository.Filter]
}

func NewDeleter(
	db dbClient.Client,
	gen *goqu.DialectWrapper,
	tableName string,
	options ...DeleterOption,
) *Deleter {
	d := &Deleter{tableName: tableName, db: db, gen: gen}

	for _, option := range options {
		option(d)
	}

	return d
}

func (d *Deleter) Delete(ctx context.Context, filters ...*repository.Filter) error {
	c := Client(d.db)
	var tx *sql.Tx
	var err error
	if d.preHooks.Count() > 0 || d.postHooks.Count() > 0 {
		tx, err = d.db.SQL().BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		c = Client(tx)
	}

	if err := d.preHooks.Fire(ctx, c, filters...); err != nil {
		return err
	}

	deleteDataset := d.gen.Delete(d.tableName).Prepared(true)

	for _, filter := range filters {
		exp, err := NewFilter(filter)
		if err != nil {
			return err
		}

		deleteDataset = deleteDataset.Where(exp)
	}

	query, args, err := deleteDataset.ToSQL()
	if err != nil {
		return err
	}

	_, err = c.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if err := d.postHooks.Fire(ctx, c, filters...); err != nil {
		return err
	}

	if tx != nil {
		return tx.Commit()
	}

	return nil
}

type DeleterOption func(u *Deleter)

func AddPreHookDeleterOption(hooks ...Hook[repository.Filter]) DeleterOption {
	return func(u *Deleter) {
		u.preHooks.RegisterHooks(hooks...)
	}
}

func AddPostHookDeleterOption(hooks ...Hook[repository.Filter]) DeleterOption {
	return func(u *Deleter) {
		u.postHooks.RegisterHooks(hooks...)
	}
}
