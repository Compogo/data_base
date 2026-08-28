package data_base

// GetIdentifier — сигнатура функции для извлечения числового идентификатора из модели.
//
// Используется компонентами Saver и Updater для определения, является ли модель
// новой (ID == 0) или уже существующей (ID != 0).
//
// Пример:
//
//	getID := func(m *User) uint64 { return m.ID }
//	id := getID(user) // 12345
//
// Обычно такой хелпер передаётся в конструктор:
//
//	saver := data_base.NewSaver(
//	    db, gen,
//	    mapper,
//	    getID,
//	    setID,
//	    "users", "id",
//	)
type GetIdentifier[T any] func(*T) uint64

// SetIdentifier — сигнатура функции для установки числового идентификатора в модель.
//
// Используется компонентами Saver и Inserter для проставления сгенерированного
// базой данных ID (например, после INSERT с auto_increment или serial колонкой).
//
// Важно: функция должна изменять переданную модель (принимает указатель),
// а не возвращать новое значение.
//
// Пример:
//
//	setID := func(m *User, id uint64) { m.ID = id }
//	setID(user, 12345)
//
// Обычно такой хелпер передаётся в конструктор:
//
//	inserter := data_base.NewInserter(
//	    db, gen,
//	    mapper,
//	    setID,
//	    "users",
//	)
type SetIdentifier[T any] func(*T, uint64)
