package GenKeys

import (
	"math/rand"
	"strings"
	"time"
)

// GenerateKey Генерация ключа путём выбора случайного символа из заданного алфавита
func GenerateKey(lenKey int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" + "_")
	var b strings.Builder
	for i := 0; i < lenKey; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

// GenerateAllKeys Генерация пула коротких ключей
func GenerateAllKeys(n, lenKey int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = GenerateKey(lenKey)
	}
	return keys
}
