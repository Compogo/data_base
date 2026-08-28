package data_base

// Identifier определяет контракт для моделей, имеющих числовой идентификатор.
// Используется компонентами Saver, Inserter и Updater для определения типа операции
// и установки сгенерированных ID.
//
// Пример реализации:
//
//	type User struct {
//	    ID uint64
//	}
//	func (u *User) GetId() uint64 { return u.ID }
//	func (u *User) SetId(id uint64) { u.ID = id }
type Identifier interface {
	GetId() uint64
	SetId(uint64)
}
