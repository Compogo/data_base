package data_base

import (
	"context"
	"database/sql"

	dbClient "github.com/Compogo/db-client"
	"github.com/Compogo/db-client/repository"
	"github.com/doug-martin/goqu/v9"
)

type BulkUpdater struct {
	tableName string
	db        dbClient.Client
	gen       *goqu.DialectWrapper

	preHooks  HookManager[repository.Filter]
	postHooks HookManager[repository.Filter]
}

func NewBulkUpdater(
	db dbClient.Client,
	gen *goqu.DialectWrapper,
	tableName string,
	options ...BulkUpdaterOption,
) *BulkUpdater {
	b := &BulkUpdater{tableName: tableName, db: db, gen: gen}

	for _, option := range options {
		option(b)
	}

	return b
}

func (b *BulkUpdater) BulkUpdate(ctx context.Context, pairs map[string]any, filters ...*repository.Filter) error {
	c := Client(b.db)
	var tx *sql.Tx
	var err error
	if b.preHooks.Count() > 0 || b.postHooks.Count() > 0 {
		tx, err = b.db.SQL().BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		c = Client(tx)
	}

	if err := b.preHooks.Fire(ctx, c, filters...); err != nil {
		return err
	}

	record := goqu.Record{}
	for columnName, value := range pairs {
		record[columnName] = value
	}

	updateDataset := b.gen.Update(record).Table(b.tableName).Prepared(true)

	for _, filter := range filters {
		exp, err := NewFilter(filter)
		if err != nil {
			return err
		}

		updateDataset = updateDataset.Where(exp)
	}

	query, args, err := updateDataset.ToSQL()
	if err != nil {
		return err
	}

	_, err = c.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if err := b.postHooks.Fire(ctx, c, filters...); err != nil {
		return err
	}

	if tx != nil {
		return tx.Commit()
	}

	return nil
}

type BulkUpdaterOption func(u *BulkUpdater)

func AddPreHookBulkUpdaterOption(hooks ...Hook[repository.Filter]) BulkUpdaterOption {
	return func(u *BulkUpdater) {
		u.preHooks.RegisterHooks(hooks...)
	}
}

func AddPostHookBulkUpdaterOption(hooks ...Hook[repository.Filter]) BulkUpdaterOption {
	return func(u *BulkUpdater) {
		u.postHooks.RegisterHooks(hooks...)
	}
}
