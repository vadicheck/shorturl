package urlservice

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/internal/repository"
	"github.com/vadicheck/shorturl/internal/services/storage/memory"
	"log"
	"os"
	"testing"
)

const userID = "4fddc63f-b1c7-48cd-b004-ff979346ea65"

func TestService_Create(t *testing.T) {
	type request struct {
		URL    string `json:"url"`
		UserID string `json:"user_id"`
	}
	tests := []struct {
		name    string
		request request
	}{
		{
			name: "Create success",
			request: request{
				URL:    "http://google.com",
				UserID: userID,
			},
		},
	}

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

	storageService, err := memory.New(tempFile.Name())
	if err != nil {
		fmt.Println("Ошибка создания хранилища:", err)
		return
	}

	ctx := context.Background()
	urlService := New(storageService)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := urlService.Create(ctx, tt.request.URL, tt.request.UserID)
			require.NoError(t, err)

			assert.Equal(t, defaultCodeLength, len(code))
		})
	}
}

func TestService_CreateBatch(t *testing.T) {
	type want []repository.BatchURL
	type request []shorten.CreateBatchURLRequest
	tests := []struct {
		name    string
		want    want
		request request
		userID  string
	}{
		{
			name: "Create success",
			want: []repository.BatchURL{
				{CorrelationID: "1", ShortCode: "43645tfr45"},
				{CorrelationID: "2", ShortCode: "gt5679jh5n"},
			},
			request: []shorten.CreateBatchURLRequest{
				{CorrelationID: "1", OriginalURL: "https://example.com/1"},
				{CorrelationID: "2", OriginalURL: "https://example.com/2"},
			},
			userID: userID,
		},
	}

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

	storageService, err := memory.New(tempFile.Name())
	if err != nil {
		fmt.Println("Ошибка создания хранилища:", err)
		return
	}

	ctx := context.Background()
	urlService := New(storageService)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities, err := urlService.CreateBatch(ctx, tt.request, tt.userID)
			require.NoError(t, err)

			assert.Equal(t, len(tt.request), len(*entities))
			assert.Equal(t, tt.request[0].CorrelationID, tt.want[0].CorrelationID)
			assert.Equal(t, tt.request[1].CorrelationID, tt.want[1].CorrelationID)

			for i := range len(tt.request) {
				assert.Equal(t, tt.request[i].CorrelationID, tt.want[i].CorrelationID)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name   string
		urls   []string
		userID string
	}{
		{
			name:   "Delete",
			urls:   []string{"https://google.com", "https://ya.ru"},
			userID: userID,
		},
	}

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

	storageService, err := memory.New(tempFile.Name())
	if err != nil {
		fmt.Println("Ошибка создания хранилища:", err)
		return
	}

	ctx := context.Background()
	urlService := New(storageService)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := urlService.Delete(ctx, tt.urls, tt.userID)
			require.NoError(t, err)
		})
	}
}

func TestService_generateCode(t *testing.T) {
	type fields struct {
		storage URLStorage
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				storage: tt.fields.storage,
			}
			got, err := s.generateCode(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateCode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
