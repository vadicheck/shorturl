package ping

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/vadicheck/shorturl/internal/services/storage/memory"
)

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

	// Инициализация хранилища.
	storage, err := memory.New(tempFile.Name())
	if err != nil {
		fmt.Println("ошибка создания хранилища:", err)
		return
	}

	// Тест успешного пинга.
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
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
	fmt.Println("Статус код (успех):", result.StatusCode)

	// Тест ошибки соединения с БД.
	errorStorage := &mockStorage{}
	reqErr := httptest.NewRequest(http.MethodGet, "/ping", nil)
	wErr := httptest.NewRecorder()

	handlerErr := New(ctx, errorStorage)
	handlerErr(wErr, reqErr)

	resultErr := wErr.Result()
	defer func() {
		if err := resultErr.Body.Close(); err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}()

	// Вывод статуса ответа в случае ошибки.
	fmt.Println("Статус код (ошибка):", resultErr.StatusCode)

	// Output:
	// Статус код (успех): 200
	// Статус код (ошибка): 500
}
