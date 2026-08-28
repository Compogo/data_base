package data_base

import (
	"context"

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
// Компонент использует функцию ModelToRecordFunc для преобразования модели в запись
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
//	updatedUser, err := updater.Update(ctx, user)
//	if err != nil {
//	    // Обработка ошибки
//	}
type Updater[T Identifier] struct {
	tableName     string
	idColName     string
	modelToRecord ModelToRecordFunc[T]
	db            dbClient.Client
	gen           *goqu.DialectWrapper
}

func NewUpdater[T Identifier](db dbClient.Client, gen *goqu.DialectWrapper, modelToRecord ModelToRecordFunc[T], tableName string, idColName string) *Updater[T] {
	return &Updater[T]{tableName: tableName, idColName: idColName, modelToRecord: modelToRecord, db: db, gen: gen}
}

func (u *Updater[T]) Update(ctx context.Context, model *T) (*T, error) {
	if (*model).GetId() == 0 {
		return nil, IdNotBeZeroError
	}

	query, args, err := u.gen.Update(u.tableName).
		Set(u.modelToRecord(model)).
		Where(goqu.C(u.idColName).Eq((*model).GetId())).
		ToSQL()

	if err != nil {
		return nil, err
	}

	_, err = u.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return model, nil
}
