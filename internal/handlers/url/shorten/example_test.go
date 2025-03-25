package shorten

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/stretchr/testify/require"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
)

const userID = "da9da41c-8f65-4ed6-abea-d58f50c41590"

type request struct {
	URL string `json:"url"`
}

// ExampleNew демонстрирует использование обработчика New.
func ExampleNew() {
	ctx := context.Background()

	// Настройка конфигурации
	config.Config.BaseURL = "http://localhost:8080"

	// Создание временного файла для хранения данных
	tempFile, err := os.CreateTemp("", "tempfile-*.json")
	if err != nil {
		fmt.Println("Ошибка создания временного файла:", err)
		return
	}
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			log.Printf("failed to remove file: %v", err)
		}
	}()

	// Инициализация хранилища
	storageService, err := memory.New(tempFile.Name())
	if err != nil {
		fmt.Println("Ошибка создания хранилища:", err)
		return
	}

	// Создание сервиса URL
	urlService := urlservice.New(storageService)

	// Создание запроса на сокращение URL

	requestBody := request{
		URL: "https://practicum.yandex.ru/",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Ошибка кодирования в JSON:", err)
		return
	}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(string(constants.XUserID), userID)

	// Создание объекта записи HTTP-ответа
	w := httptest.NewRecorder()

	// Вызов обработчика
	handler := New(ctx, urlService)
	handler(w, req)

	// Получение результата
	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}()

	// Проверка кода состояния
	fmt.Println("Статус код:", result.StatusCode)

	// Декодирование JSON-ответа
	var response shorten.CreateURLResponse
	decoder := json.NewDecoder(result.Body)
	err = decoder.Decode(&response)
	require.NoError(nil, err)

	response.Result = "http://localhost:8080/abc123"

	// Вывод результата
	fmt.Println("Ответ:", response)

	// Output:
	// Статус код: 201
	// Ответ: {http://localhost:8080/abc123}
}
