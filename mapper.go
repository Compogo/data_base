package data_base

import (
	"github.com/doug-martin/goqu/v9"
)

// Mapper — сигнатура функции для преобразования указателя на модель,
// используется в запись для SQL-запроса (goqu.Record). Используется компонентами Saver, Inserter, Updater и BulkInserter.
//
// Позволяет гибко контролировать, какие поля и как сохраняются в базе данных.
type Mapper[T any] func(*T) goqu.Record
