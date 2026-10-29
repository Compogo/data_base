package data_base

import (
	"context"
	"database/sql"

	dbClient "github.com/Compogo/db-client"
	"github.com/doug-martin/goqu/v9"
)

// Updater предназначен для обновления существующих записей в таблице.
// В отличие от Saver, Updater не выполняет вставку новых записей и требует,
// чтобы переданная модель имела заполненный идентификатор (GetId() != 0).
//
// Если идентификатор модели равен 0, метод Update возвращает ошибку IdNotBeZeroError.
// Это предотвращает случайное выполнение UPDATE без условия WHERE.
//
// Компонент использует функцию Mapper для преобразования модели в запись
// (goqu.Record) и выполняет обновление по указанному полю идентификатора.
//
// Пример использования:
//
//	toRecord := func(u *User) goqu.Record {
//	    return goqu.Record{"name": u.Name, "email": u.Email}
//	}
//
//	updater := data_base.NewUpdater(
//	    db,
//	    goqu.Dialect("postgres"),
//	    toRecord,
//	    "users",
//	    "id",
//	)
//
//	user := &User{Id: 123, Name: "Updated Name"}
//	updatedUser, err := updater.Update[User](ctx, user)
//	if err != nil {
//	    // Обработка ошибки
//	}
type Updater[T any] struct {
	tableName     string
	idColName     string
	mapper        Mapper[T]
	getIdentifier GetIdentifier[T]
	db            dbClient.Client
	gen           *goqu.DialectWrapper

	preHooks  HookManager[T]
	postHooks HookManager[T]
}

func NewUpdater[T any](
	db dbClient.Client,
	gen *goqu.DialectWrapper,
	mapper Mapper[T],
	getIdentifier GetIdentifier[T],
	tableName string,
	idColName string,
	options ...UpdaterOption[T],
) *Updater[T] {
	r := &Updater[T]{
		tableName:     tableName,
		idColName:     idColName,
		mapper:        mapper,
		db:            db,
		gen:           gen,
		getIdentifier: getIdentifier,
	}

	for _, option := range options {
		option(r)
	}

	return r
}

func (u *Updater[T]) Update(ctx context.Context, model *T) (*T, error) {
	if u.getIdentifier(model) == 0 {
		return nil, IdNotBeZeroError
	}

	c := Client(u.db)
	var tx *sql.Tx
	var err error
	if u.preHooks.Count() > 0 || u.postHooks.Count() > 0 {
		tx, err = u.db.SQL().BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		c = Client(tx)
	}

	if err := u.preHooks.Fire(ctx, c, model); err != nil {
		return nil, err
	}

	query, args, err := u.gen.Update(u.tableName).
		Set(u.mapper(model)).
		Where(goqu.C(u.idColName).Eq(u.getIdentifier(model))).
		Prepared(true).
		ToSQL()

	if err != nil {
		return nil, err
	}

	_, err = c.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if err := u.postHooks.Fire(ctx, c, model); err != nil {
		return nil, err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}

	return model, nil
}

type UpdaterOption[T any] func(u *Updater[T])

func AddPreHookUpdaterOption[T any](hooks ...Hook[T]) UpdaterOption[T] {
	return func(u *Updater[T]) {
		u.preHooks.RegisterHooks(hooks...)
	}
}

func AddPostHookUpdaterOption[T any](hooks ...Hook[T]) UpdaterOption[T] {
	return func(u *Updater[T]) {
		u.postHooks.RegisterHooks(hooks...)
	}
}
