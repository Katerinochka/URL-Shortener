package IO

import (
	"URL-Shortener/internal/GenKeys"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"log"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(nKeys, lenKey int) (*Postgres, error) {
	var err error
	postgres := new(Postgres)
	//postgres.db, err = sql.Open("postgres", "user=postgres dbname=postgres sslmode=disable")
	//postgres.db, err = sql.Open("postgres", "postgres://postgres:postgres@db:5432/postgres?sslmode=disable")
	postgres.db, err = sql.Open("postgres", "host=postgres port=5432 user=postgres password=qwerty dbname=postgres sslmode=disable")
	if err != nil {
		return nil, err
	}

	// При новом запуске очищаем БД от старых записей
	postgres.db.Exec("DROP TABLE IF EXISTS freekeys")
	postgres.db.Exec("CREATE TABLE freekeys (short varchar(10) NOT NULL)")
	postgres.db.Exec("DROP TABLE IF EXISTS busykeys")
	postgres.db.Exec("CREATE TABLE busykeys (short varchar(10) NOT NULL, origurl text NOT NULL, created_at timestamp DEFAULT NOW())")

	// Генерим пул свободных ключей
	keys := GenKeys.GenerateAllKeys(nKeys, lenKey)

	// Заполняем таблицу freekeys. если на этом этапе произошла ошибка, то сервер останавливается
	//https://pkg.go.dev/github.com/lib/pq#hdr-Bulk_imports
	txn, err := postgres.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("freekeys", "short"))
	if err != nil {
		log.Fatal(err)
	}

	for _, key := range keys {
		_, err = stmt.Exec(key)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = txn.Commit()
	if err != nil {
		log.Fatal(err)
	}

	return postgres, nil
}

// PushFreeKeys Вспомогательная функция генерации одного короткого ключа
func (postgres *Postgres) PushFreeKeys(lenKey int) {
	postgres.db.QueryRow(
		"INSERT INTO freekeys (short) VALUES ($1)",
		GenKeys.GenerateKey(lenKey),
	).Scan()
}

// FrontFreeKeys Отдаёт из пула сгенерированных коротких ключей один ключ, добавляет в конец пула новый сгенерированный ключ
func (postgres *Postgres) FrontFreeKeys(lenKey int) (string, error) {
	var key string
	// Забираем верхний ключ из пула свободных
	err := postgres.db.QueryRow("SELECT * FROM freekeys LIMIT 1").Scan(&key)

	// Если ключи закончились, просим пользователя повторить попытку
	if err != nil {
		return "", errors.New("no free keys found, please try again")
	}

	// Удаляем этот ключ из пула
	postgres.db.QueryRow(
		"DELETE FROM freekeys WHERE short= $1",
		key).Scan()

	// На его место генерим новый
	postgres.PushFreeKeys(lenKey)
	return key, nil
}

// PushBusyKeys Добавляем короткий ключ, оригинальную ссылку и время создания в базу с занятыми ключами
func (postgres *Postgres) PushBusyKeys(short, long string) error {
	var busyShort string
	// Если пытаемся добавить в занятые ключ, который уже в ней есть, значит взяли из пула дубликат, просим пользователя повторить попытку
	postgres.db.QueryRow(
		"SELECT short FROM busykeys WHERE short = $1",
		short).Scan(&busyShort)

	if busyShort != "" {
		return errors.New("a non-unique key was generated, please try again")
	}

	// Добавляем в пул занятых ключей
	postgres.db.QueryRow(
		"INSERT INTO busykeys (short, origurl) VALUES ($1, $2)",
		short,
		long,
	).Scan()

	return nil
}

// Find Ищем по ключу короткой ссылки оригинальную
func (postgres *Postgres) Find(shortKey string) (string, error) {
	var origUrl string
	postgres.db.QueryRow(
		"SELECT origurl FROM busykeys WHERE short = $1",
		shortKey,
	).Scan(&origUrl)

	// Если не нашли, говорим об этом пользователю
	if origUrl == "" {
		return "", errors.New("Origin link not found")
	}

	return origUrl, nil
}

// CheckExistingOriginal Проверяем наличие оригинальной ссылки
func (postgres *Postgres) CheckExistingOriginal(long string) (string, error) {
	var short string
	postgres.db.QueryRow(
		"SELECT short FROM busykeys WHERE origurl = $1",
		long,
	).Scan(&short)

	if short != "" {
		return short, errors.New("Ok. Return existing link.")
	}

	return "", nil
}

func (postgres *Postgres) Cleaning() {
	postgres.db.QueryRow(
		"DELETE FROM busykeys WHERE CURRENT_TIMESTAMP - created_at > INTERVAL '1 day'",
	).Scan()
}
