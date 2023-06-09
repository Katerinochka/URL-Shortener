package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	ShortLinkTitle = "http://short.com/"
)

type LongRequest struct {
	Long string `json:"longURL"`
}

type ShortResponse struct {
	StatusMessage string `json:"statusmessage""`
	Short         string `json:"shortURL"`
}

type LongResponse struct {
	StatusMessage string `json:"statusmessage"`
	Long          string `json:"longURL"`
}

// Обработка POST запроса пользователя (приходит оригинальная ссылка, отправляется короткая и сообщение)
func CreateShortURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	//Извлекаем json с оригинальной ссылкой
	requestDecoder := json.NewDecoder(r.Body)
	responseEncoder := json.NewEncoder(w)
	urlRequest := new(LongRequest)
	err := requestDecoder.Decode(&urlRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err := responseEncoder.Encode(&ShortResponse{StatusMessage: err.Error()}); err != nil {
			fmt.Fprintf(w, "Request processing error %v\n", err.Error())
		}
		return
	}

	// Если оригинальная ссылка была пустая, отправляем в ответ соответствующую ошибку
	if len(urlRequest.Long) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		if err := responseEncoder.Encode(&ShortResponse{StatusMessage: "Empty line."}); err != nil {
			fmt.Fprintf(w, "Request processing error %v\n", err.Error())
		}
		return
	}

	// Если для оригинальной ссылки уже была сгенерирована короткая, то отправляем её, а не создаём новую
	if short, err := io.CheckExistingOriginal(urlRequest.Long); err != nil {
		if err = responseEncoder.Encode(&ShortResponse{StatusMessage: err.Error(), Short: ShortLinkTitle + short}); err != nil {
			fmt.Fprintf(w, "Request processing error %v\n", err.Error())
		}
		return
	}

	// Если генерируем ссылку впервые, берём её из уже сгенерированного пула ключей
	short, err := io.FrontFreeKeys(LenKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := responseEncoder.Encode(&ShortResponse{StatusMessage: err.Error()}); err != nil {
			fmt.Fprintf(w, "Request processing error %v\n", err.Error())
		}
		return
	}

	// Добавляем короткий ключ в таблицу занятых ключей
	err = io.PushBusyKeys(short, urlRequest.Long)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := responseEncoder.Encode(&ShortResponse{StatusMessage: err.Error()}); err != nil {
			fmt.Fprintf(w, "Request processing error %v\n", err.Error())
		}
		return
	}

	// Если всё сложилось удачно (короткая уникальная ссылка сгенерирована), отправяем короткую ссылку пользователю
	if err = responseEncoder.Encode(&ShortResponse{StatusMessage: "Ok", Short: ShortLinkTitle + short}); err != nil {
		fmt.Fprintf(w, "Request processing error %v\n", err.Error())
	}
}

// Обработка GET запроса пользователя (приходит в параметре короткая ссылка, отправляет оригинальная и сообщение)
func GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	// Получаем из запроса короткую ссылку
	shortURL := r.URL.Query().Get("shorturl")
	shortKey, found := strings.CutPrefix(shortURL, ShortLinkTitle)
	responseEncoder := json.NewEncoder(w)
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		if err := responseEncoder.Encode(&ShortResponse{StatusMessage: "wrong url prefix"}); err != nil {
			fmt.Fprintf(w, "Request processing error %v\n", err.Error())
		}
		return
	}

	// Ищем в хранилище соответствующую оригинальную
	originUrl, err := io.Find(shortKey)
	// Если оригинальная не найдена по короткой, отправляем соответствующее сообщение
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err := responseEncoder.Encode(&ShortResponse{StatusMessage: "Origin link not found"}); err != nil {
			fmt.Fprintf(w, "Request processing error %v\n", err.Error())
		}
		return
	}

	// Если все сложилось удачно (оригинальная ссылка найдена), отправляем пользователю оригинальную ссылку
	if err := responseEncoder.Encode(&LongResponse{StatusMessage: "Ok", Long: originUrl}); err != nil {
		fmt.Fprintf(w, "Request processing error %v\n", err.Error())
	}
}
