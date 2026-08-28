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
//	inserter := data_base.NewInserter(db, gen, toRecord, "users")
//	user := &User{Name: "Bob", Email: "bob@example.com"}
//	insertedUser, err := inserter.Insert(ctx, user)
type Inserter[T Identifier] struct {
	tableName     string
	modelToRecord ModelToRecordFunc[T]
	db            dbClient.Client
	gen           *goqu.DialectWrapper
}

func NewInserter[T Identifier](db dbClient.Client, gen *goqu.DialectWrapper, modelToRecord ModelToRecordFunc[T], tableName string) *Inserter[T] {
	return &Inserter[T]{tableName: tableName, modelToRecord: modelToRecord, db: db, gen: gen}
}

func (i *Inserter[T]) Insert(ctx context.Context, model *T) (*T, error) {
	query, args, err := i.gen.Insert(i.tableName).Rows(i.modelToRecord(model)).Prepared(true).ToSQL()
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

	(*model).SetId(uint64(id))

	return model, nil
}
