package data_base

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
)

// RowToModelFunc — сигнатура функции для преобразования строки результата SQL-запроса (*sql.Rows),
// используется в экземпляр модели типа T. Используется компонентами Finder и Pager.
//
// Возвращает указатель на модель и ошибку, если сканирование не удалось.
type RowToModelFunc[T any] func(*sql.Rows) (*T, error)

// ModelToRecordFunc — сигнатура функции для преобразования указателя на модель,
// используется в запись для SQL-запроса (goqu.Record). Используется компонентами Saver, Inserter, Updater и BulkInserter.
//
// Позволяет гибко контролировать, какие поля и как сохраняются в базе данных.
type ModelToRecordFunc[T any] func(*T) goqu.Record
