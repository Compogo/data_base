package data_base

import (
	"context"

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
}

func (b *BulkInserter[T]) BulkInsert(ctx context.Context, models ...*T) error {
	rows := make([]any, len(models), 0)
	for _, model := range models {
		rows = append(rows, b.modelToRecord(model))
	}

	query, args, err := b.gen.Insert(b.tableName).Rows(rows).Prepared(true).ToSQL()
	if err != nil {
		return err
	}

	_, err = b.db.ExecContext(ctx, query, args...)
	return err
}
