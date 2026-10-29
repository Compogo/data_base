package data_base

import (
	"context"
	"database/sql"

	dbClient "github.com/Compogo/db-client"
	"github.com/doug-martin/goqu/v9"
)

// Inserter предназначен только для вставки новых записей (INSERT).
// После успешной вставки автоматически устанавливает сгенерированный ID модели через SetId().
//
// В отличие от Saver, не выполняет обновление при наличии ID.
// Полезен для сценариев, где нужно гарантировать создание новой записи.
//
// Пример:
//
//	inserter := data_base.NewInserter[User](db, gen, toRecord, "users")
//	user := &User{Name: "Bob", Email: "bob@example.com"}
//	insertedUser, err := inserter.Insert(ctx, user)
type Inserter[T any] struct {
	tableName     string
	mapper        Mapper[T]
	setIdentifier SetIdentifier[T]
	db            dbClient.Client
	gen           *goqu.DialectWrapper

	preHooks  HookManager[T]
	postHooks HookManager[T]
}

func NewInserter[T any](
	db dbClient.Client,
	gen *goqu.DialectWrapper,
	mapper Mapper[T],
	setIdentifier SetIdentifier[T],
	tableName string,
	options ...InserterOption[T],
) *Inserter[T] {
	i := &Inserter[T]{
		tableName:     tableName,
		mapper:        mapper,
		db:            db,
		gen:           gen,
		setIdentifier: setIdentifier,
	}

	for _, option := range options {
		option(i)
	}

	return i
}

func (i *Inserter[T]) Insert(ctx context.Context, model *T) (*T, error) {
	c := Client(i.db)
	var tx *sql.Tx
	var err error
	if i.preHooks.Count() > 0 || i.postHooks.Count() > 0 {
		tx, err = i.db.SQL().BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		c = Client(tx)
	}

	if err := i.preHooks.Fire(ctx, c, model); err != nil {
		return nil, err
	}

	query, args, err := i.gen.Insert(i.tableName).Rows(i.mapper(model)).Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	result, err := c.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	i.setIdentifier(model, uint64(id))

	if err := i.postHooks.Fire(ctx, c, model); err != nil {
		return nil, err
	}

	return model, nil
}

type InserterOption[T any] func(u *Inserter[T])

func AddPreHookInserterOption[T any](hooks ...Hook[T]) InserterOption[T] {
	return func(u *Inserter[T]) {
		u.preHooks.RegisterHooks(hooks...)
	}
}

func AddPostHookInserterOption[T any](hooks ...Hook[T]) InserterOption[T] {
	return func(u *Inserter[T]) {
		u.postHooks.RegisterHooks(hooks...)
	}
}
