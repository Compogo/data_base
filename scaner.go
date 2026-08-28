package data_base

import "database/sql"

// Scanner — сигнатура функции для преобразования строки результата SQL-запроса (*sql.Rows),
// используется в экземпляр модели типа T. Используется компонентами Finder и Pager.
//
// Возвращает указатель на модель и ошибку, если сканирование не удалось.
type Scanner[T any] func(*sql.Rows) (*T, error)
