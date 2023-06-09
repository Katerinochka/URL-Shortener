package IO

import (
	"URL-Shortener/internal/GenKeys"
	"errors"
	"time"
)

type freeKeys []string

type busyKey struct {
	LongUrl string
	Time    time.Time
}

type InMemory struct {
	FreeKeys     freeKeys           // Список заранее сгенерированных свободных ключей
	BusyKeys     map[string]busyKey // Мапа с занятыми ключами. Ключ - короткая ссылка, значение - структура с оригинальной ссылкой и временем создания
	OriginalUrls map[string]string  // Мапа с оригинальными ключами. Для проверки уникальности оригинальной ссылки. Ключ - оригинальный url, значение - короткий ключ
}

func NewInOut(nKeys, lenKey int) *InMemory {
	im := new(InMemory)
	// Генерим пул свободных ключей при запуске
	im.FreeKeys = GenKeys.GenerateAllKeys(nKeys, lenKey)
	im.BusyKeys = make(map[string]busyKey)
	im.OriginalUrls = make(map[string]string)
	return im
}

// PushFreeKeys Вспомогательная функция генерации одного короткого ключа
func (im *InMemory) PushFreeKeys(lenKey int) {
	im.FreeKeys = append(im.FreeKeys, GenKeys.GenerateKey(lenKey))
}

// FrontFreeKeys Отдаёт из пула сгенерированных коротких ключей один ключ, добавляет в конец пула новый сгенерированный ключ
func (im *InMemory) FrontFreeKeys(lenKey int) (string, error) {
	// Если ключи закончились, просим пользователя повторить попытку
	if len(im.FreeKeys) == 0 {
		return "", errors.New("no free keys found, please try again")
	}
	frontKey := im.FreeKeys[0]
	im.FreeKeys = im.FreeKeys[1:]
	im.PushFreeKeys(lenKey)
	return frontKey, nil
}

// PushBusyKeys Добавляем короткий ключ, оригинальную ссылку и время создания в базу с занятыми ключами
func (im *InMemory) PushBusyKeys(short, long string) error {
	// Если пытаемся добавить в занятые ключ, который уже в ней есть, значит взяли из пула дубликат, просим пользователя повторить попытку
	if _, ok := im.BusyKeys[short]; ok {
		return errors.New("a non-unique key was generated, please try again")
	}
	// Добавляем в пул занятых ключей
	im.BusyKeys[short] = busyKey{
		LongUrl: long,
		Time:    time.Now(),
	}
	// Добавляем наличие оригинальной ссылки
	im.OriginalUrls[long] = short
	return nil
}

// Find Ищем по ключу короткой ссылки оригинальную
func (im *InMemory) Find(shortKey string) (string, error) {
	if key, ok := im.BusyKeys[shortKey]; ok {
		return key.LongUrl, nil
	}
	// Если не нашли, говорим об этом пользователю
	return "", errors.New("Origin link not found")
}

// CheckExistingOriginal Проверяем наличие оригинальной ссылки
func (im *InMemory) CheckExistingOriginal(long string) (string, error) {
	if short, ok := im.OriginalUrls[long]; ok {
		return short, errors.New("Ok. Return existing link.")
	}
	return "", nil
}

func (im *InMemory) Cleaning() {
	tNow := time.Now()
	for short, bKey := range im.BusyKeys {
		if tNow.Sub(bKey.Time) > 24*time.Hour {
			delete(im.BusyKeys, short)
		}
	}
}
