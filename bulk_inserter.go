package data_base

import (
	"context"
	"database/sql"

	dbClient "github.com/Compogo/db-client"
	"github.com/doug-martin/goqu/v9"
)

// BulkInserter предназначен для массовой вставки записей (INSERT) одной операцией.
// Принимает срез указателей на модели и преобразует их в записи с помощью Mapper.
//
// Значительно повышает производительность при вставке большого количества записей,
// сокращая количество сетевых вызовов к базе данных.
//
// Пример:
//
//	bulk := &data_base.BulkInserter[User]{
//	    tableName: "users",
//	    mapper: toRecord,
//	    db: db,
//	    gen: gen,
//	}
//	users := []*User{&User{Name: "A"}, &User{Name: "B"}}
//	err := bulk.BulkInsert(ctx, users...)
type BulkInserter[T any] struct {
	tableName     string
	modelToRecord Mapper[T]
	db            dbClient.Client
	gen           *goqu.DialectWrapper

	preHooks  HookManager[T]
	postHooks HookManager[T]
}

func NewBulkInserter[T any](
	db dbClient.Client,
	gen *goqu.DialectWrapper,
	modelToRecord Mapper[T],
	tableName string,
	options ...BulkInserterOption[T],
) *BulkInserter[T] {
	b := &BulkInserter[T]{tableName: tableName, modelToRecord: modelToRecord, db: db, gen: gen}

	for _, option := range options {
		option(b)
	}

	return b
}

func (b *BulkInserter[T]) BulkInsert(ctx context.Context, models ...*T) error {
	rows := make([]any, len(models), 0)
	for _, model := range models {
		rows = append(rows, b.modelToRecord(model))
	}

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

	if err := b.preHooks.Fire(ctx, c, models...); err != nil {
		return err
	}

	query, args, err := b.gen.Insert(b.tableName).Rows(rows).Prepared(true).ToSQL()
	if err != nil {
		return err
	}

	_, err = c.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if err := b.postHooks.Fire(ctx, c, models...); err != nil {
		return err
	}

	if tx != nil {
		return tx.Commit()
	}

	return nil
}

type BulkInserterOption[T any] func(u *BulkInserter[T])

func AddPreHookBulkInserterOption[T any](hooks ...Hook[T]) BulkInserterOption[T] {
	return func(u *BulkInserter[T]) {
		u.preHooks.RegisterHooks(hooks...)
	}
}

func AddPostHookBulkInserterOption[T any](hooks ...Hook[T]) BulkInserterOption[T] {
	return func(u *BulkInserter[T]) {
		u.postHooks.RegisterHooks(hooks...)
	}
}
