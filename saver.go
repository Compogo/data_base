package data_base

import (
	"context"
	"database/sql"

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

	preHooks  HookManager[T]
	postHooks HookManager[T]
}

func NewSaver[T any](
	db dbClient.Client,
	gen *goqu.DialectWrapper,
	mapper Mapper[T],
	getIdentifier GetIdentifier[T],
	setIdentifier SetIdentifier[T],
	tableName string,
	idColName string,
	options ...SaverOption[T],
) *Saver[T] {
	s := &Saver[T]{
		tableName:     tableName,
		idColName:     idColName,
		mapper:        mapper,
		db:            db,
		gen:           gen,
		getIdentifier: getIdentifier,
		setIdentifier: setIdentifier,
	}

	for _, option := range options {
		option(s)
	}

	return s
}

func (s *Saver[T]) Save(ctx context.Context, model *T) (*T, error) {
	c := Client(s.db)
	var tx *sql.Tx
	var err error
	if s.preHooks.Count() > 0 || s.postHooks.Count() > 0 {
		tx, err = s.db.SQL().BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		c = Client(tx)
	}

	if err := s.preHooks.Fire(ctx, c, model); err != nil {
		return nil, err
	}

	record := s.mapper(model)

	var queryBuilder exp.SQLExpression

	if s.getIdentifier(model) == 0 {
		queryBuilder = s.gen.Insert(s.tableName).Rows(record).Prepared(true)
	} else {
		queryBuilder = s.gen.Update(s.tableName).Set(record).Where(goqu.C(s.idColName).Eq(s.getIdentifier(model))).Prepared(true)
	}

	query, args, err := queryBuilder.ToSQL()
	if err != nil {
		return nil, err
	}

	result, err := c.ExecContext(ctx, query, args...)
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

	if err := s.postHooks.Fire(ctx, c, model); err != nil {
		return nil, err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}

	return model, nil
}

type SaverOption[T any] func(u *Saver[T])

func AddPreHookSaverOption[T any](hooks ...Hook[T]) SaverOption[T] {
	return func(u *Saver[T]) {
		u.preHooks.RegisterHooks(hooks...)
	}
}

func AddPostHookSaverOption[T any](hooks ...Hook[T]) SaverOption[T] {
	return func(u *Saver[T]) {
		u.postHooks.RegisterHooks(hooks...)
	}
}
