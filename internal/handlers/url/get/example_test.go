package get

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
)

// mockStorage имитирует хранилище с разными сценариями.
type mockStorage struct {
	data map[string]models.URL
}

func (m *mockStorage) GetURLByID(ctx context.Context, code string) (models.URL, error) {
	url, exists := m.data[code]
	if !exists {
		return models.URL{}, nil
	}
	return url, nil
}

// ExampleNew демонстрирует использование обработчика New.
func ExampleNew() {
	ctx := context.Background()

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

	// Инициализация реального хранилища.
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

	// Тест успешного получения URL.
	req := httptest.NewRequest(http.MethodGet, "/{id}", http.NoBody)
	req.SetPathValue("id", "example")
	w := httptest.NewRecorder()

	handler := New(ctx, storage)
	handler(w, req)

	result := w.Result()
	defer func() {
		if err := result.Body.Close(); err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}()

	// Вывод статуса ответа
	fmt.Println("Статус код (существующий URL):", result.StatusCode)
	fmt.Println("Location:", result.Header.Get("Location"))

	// Тест отсутствующего URL.
	mock := &mockStorage{data: map[string]models.URL{}}
	reqNotFound := httptest.NewRequest(http.MethodGet, "/{id}", http.NoBody)
	reqNotFound.SetPathValue("id", "notfound")
	wNotFound := httptest.NewRecorder()

	handlerNotFound := New(ctx, mock)
	handlerNotFound(wNotFound, reqNotFound)

	resultNotFound := wNotFound.Result()
	defer func() {
		if err := resultNotFound.Body.Close(); err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}()

	// Вывод статуса ответа для отсутствующего URL.
	fmt.Println("Статус код (отсутствующий URL):", resultNotFound.StatusCode)

	// Тест удалённого URL.
	mockDeleted := &mockStorage{
		data: map[string]models.URL{
			"deleted": {ID: 2, Code: "deleted", URL: "https://deleted.com/", IsDeleted: true},
		},
	}
	reqDeleted := httptest.NewRequest(http.MethodGet, "/{id}", http.NoBody)
	reqDeleted.SetPathValue("id", "deleted")
	wDeleted := httptest.NewRecorder()

	handlerDeleted := New(ctx, mockDeleted)
	handlerDeleted(wDeleted, reqDeleted)

	resultDeleted := wDeleted.Result()
	defer func() {
		if err := resultDeleted.Body.Close(); err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}()

	// Вывод статуса ответа для удалённого URL.
	fmt.Println("Статус код (удалённый URL):", resultDeleted.StatusCode)

	// Output:
	// Статус код (существующий URL): 307
	// Location: https://example.com/
	// Статус код (отсутствующий URL): 404
	// Статус код (удалённый URL): 410
}
