package urls

import (
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
	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
)

const userID = "da9da41c-8f65-4ed6-abea-d58f50c41590"

// ExampleNew демонстрирует использование обработчика New.
func ExampleNew() {
	ctx := context.Background()

	// Настройка конфигурации
	config.Config.BaseURL = "http://localhost:8080"

	// Создание временного файла для хранения данных.
	tempFile, err := os.CreateTemp("", "tempfile-*.json")
	if err != nil {
		fmt.Println("ошибка создания временного файла:", err)
		return
	}
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			log.Printf("failed to remove file: %v", err)
		}
	}()

	// Инициализация хранилища.
	storage, err := memory.New(tempFile.Name())
	if err != nil {
		fmt.Println("ошибка создания хранилища:", err)
		return
	}

	// Добавление тестовых данных.
	testURL := models.URL{
		Code:   "example",
		URL:    "https://example.com/",
		UserID: userID,
	}
	_, err = storage.SaveURL(ctx, testURL.Code, testURL.URL, testURL.UserID)
	if err != nil {
		fmt.Println("ошибка сохранения URL:", err)
		return
	}

	// Создание запроса.
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req.Header.Set(string(constants.XUserID), userID)

	// Создание объекта записи HTTP-ответа.
	w := httptest.NewRecorder()

	// Вызов обработчика.
	handler := New(ctx, storage)
	handler(w, req)

	// Получение результата.
	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}()

	// Проверка кода состояния.
	fmt.Println("Статус код:", result.StatusCode)

	// Декодирование JSON-ответа.
	var response []shorten.UserURLResponse
	decoder := json.NewDecoder(result.Body)
	err = decoder.Decode(&response)
	require.NoError(nil, err)

	// Вывод результата.
	fmt.Println("Ответ:", response)

	// Output:
	// Статус код: 200
	// Ответ: [{http://localhost:8080/example https://example.com/}]
}
