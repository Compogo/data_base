package data_base

import (
	"context"

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
}

func NewInserter[T any](
	db dbClient.Client,
	gen *goqu.DialectWrapper,
	mapper Mapper[T],
	setIdentifier SetIdentifier[T],
	tableName string,
) *Inserter[T] {
	return &Inserter[T]{
		tableName:     tableName,
		mapper:        mapper,
		db:            db,
		gen:           gen,
		setIdentifier: setIdentifier,
	}
}

func (i *Inserter[T]) Insert(ctx context.Context, model *T) (*T, error) {
	query, args, err := i.gen.Insert(i.tableName).Rows(i.mapper(model)).Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	result, err := i.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	i.setIdentifier(model, uint64(id))

	return model, nil
}
