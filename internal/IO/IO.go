package IO

// Интерфейс ввода/вывода
type InOut interface {
	FrontFreeKeys(lenKey int) (string, error)
	PushBusyKeys(short, long string) error
	Find(shortKey string) (string, error)
	CheckExistingOriginal(long string) (string, error)
	Cleaning()
}
