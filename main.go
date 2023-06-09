package main

import (
	"URL-Shortener/internal/IO"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var io IO.InOut

const (
	LenKey = 10
	NKeys  = 1000000
)

func init() {
	var flagIO string
	// Считываем флаг - режим хранилища
	flag.StringVar(&flagIO, "storage", "inmem", "storage mode: inmem or db")
	flag.Parse()

	// Во время создания нужного хранилища, происходит генерация пула свободных ключей
	fmt.Println("Wait, key generation in progress")

	switch flagIO {
	case "inmem":
		io = IO.NewInOut(NKeys, LenKey)
	case "db":
		io, _ = IO.NewPostgres(NKeys, LenKey)
	default:
		log.Fatal("storage mode: inmem or db")
	}
	fmt.Println("Key generation completed")
}

func main() {
	go func() {
		for {
			timeNow := time.Now()
			if timeNow.Hour() == 0 && timeNow.Minute() == 0 && timeNow.Second() == 0 {
				io.Cleaning()
				time.Sleep(23 * time.Hour)
			}
		}
	}()
	http.HandleFunc("/create", CreateShortURL)
	http.HandleFunc("/getoriginal", GetOriginalURL)
	http.ListenAndServe(":8080", nil)
}
