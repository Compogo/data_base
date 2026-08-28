package data_base

import (
	"context"

	dbClient "github.com/Compogo/db-client"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// Saver реализует паттерн "сохранить или обновить" (UPSERT), автоматически
// определяя операцию по значению идентификатора модели.
//
// Если GetId() возвращает 0, выполняется вставка (INSERT) новой записи,
// после чего модели автоматически устанавливается сгенерированный ID через SetId().
// Если GetId() не равен 0, выполняется обновление (UPDATE) существующей записи.
//
// Требует, чтобы тип T реализовывал интерфейс Identifier.
// Для работы необходима функция, преобразующая модель в goqu.Record.
//
// Пример:
//
//	toRecord := func(u *User) goqu.Record {
//	    return goqu.Record{"name": u.Name, "email": u.Email}
//	}
//
//	saver := data_base.NewSaver[User](db, gen, toRecord, "users", "id")
//	user := &User{Name: "Alice", Email: "alice@example.com"}
//	savedUser, err := saver.Save(ctx, user) // user.ID будет установлен после вставки
type Saver[T any] struct {
	tableName     string
	idColName     string
	mapper        Mapper[T]
	getIdentifier GetIdentifier[T]
	setIdentifier SetIdentifier[T]
	db            dbClient.Client
	gen           *goqu.DialectWrapper
}

func NewSaver[T any](
	db dbClient.Client,
	gen *goqu.DialectWrapper,
	mapper Mapper[T],
	getIdentifier GetIdentifier[T],
	setIdentifier SetIdentifier[T],
	tableName string,
	idColName string,
) *Saver[T] {
	return &Saver[T]{
		tableName:     tableName,
		idColName:     idColName,
		mapper:        mapper,
		db:            db,
		gen:           gen,
		getIdentifier: getIdentifier,
		setIdentifier: setIdentifier,
	}
}

func (s *Saver[T]) Save(ctx context.Context, model *T) (*T, error) {
	record := s.mapper(model)

	var queryBuilder exp.SQLExpression

	if s.getIdentifier(model) == 0 {
		queryBuilder = s.gen.Insert(s.tableName).Rows(record).Prepared(true)
	} else {
		queryBuilder = s.gen.Update(s.tableName).Set(record).Where(goqu.C(s.idColName).Eq(s.getIdentifier(model)))
	}

	query, args, err := queryBuilder.ToSQL()
	if err != nil {
		return nil, err
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if s.getIdentifier(model) == 0 {
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}

		s.setIdentifier(model, uint64(id))
	}

	return model, nil
}
