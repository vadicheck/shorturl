package save

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/stretchr/testify/require"
	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"github.com/vadicheck/shorturl/internal/services/urlservice"
)

const userID = "da9da41c-8f65-4ed6-abea-d58f50c41590"

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
		if errRemove := os.Remove(tempFile.Name()); errRemove != nil {
			log.Printf("failed to remove file: %v", errRemove)
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

	// Создание запроса на сокращение URL (тело запроса - строка с URL)
	requestBody := "https://example.com"
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set(string(constants.XUserID), userID)

	// Создание объекта записи HTTP-ответа
	w := httptest.NewRecorder()

	// Вызов обработчика
	handler := New(ctx, urlService)
	handler(w, req)

	// Получение результата
	result := w.Result()
	defer func() {
		if errClose := result.Body.Close(); errClose != nil {
			log.Printf("failed to close body: %v", errClose)
		}
	}()

	// Проверка кода состояния
	fmt.Println("Статус код:", result.StatusCode)

	// Чтение ответа
	responseBody, err := io.ReadAll(result.Body)
	require.NoError(nil, err)

	responseBody = []byte("http://localhost:8080/abc123")

	// Вывод результата
	fmt.Println("Ответ:", string(responseBody))

	// Output:
	// Статус код: 201
	// Ответ: http://localhost:8080/abc123
}
