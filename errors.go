package data_base

import "errors"

// IdNotBeZeroError возвращается компонентом Updater (и другими операциями обновления),
// если переданная модель имеет нулевой идентификатор (GetId() == 0).
//
// Такая проверка позволяет избежать случайного обновления всех записей в таблице
// (например, при WHERE id = 0) и даёт явный сигнал о некорректном вызове.
//
// Пример:
//
//	user := &User{Name: "Alice"} // ID не установлен
//	_, err := updater.Update(ctx, user)
//	if errors.Is(err, data_base.IdNotBeZeroError) {
//	    // Обработка ошибки: ID обязателен для обновления
//	}
var IdNotBeZeroError = errors.New("id must not be null")
