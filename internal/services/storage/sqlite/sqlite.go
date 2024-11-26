package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/services/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", storagePath, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUrl(ctx context.Context, code string, url string) (int64, error) {
	const op = "storage.sqlite.SaveUrl"

	stmt, err := s.db.Prepare("INSERT INTO main.urls (code, url) VALUES (?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, code, url)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUrlOrCodeExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUrlById(ctx context.Context, code string) (models.Url, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, code, url FROM urls WHERE code=?", code)

	return s.scan(row, "storage.sqlite.GetUrlById")
}

func (s *Storage) GetUrlByUrl(ctx context.Context, url string) (models.Url, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, code, url FROM urls WHERE url=?", url)

	return s.scan(row, "storage.sqlite.GetUrlByUrl")
}

func (s *Storage) scan(row *sql.Row, op string) (models.Url, error) {
	var modelUrl models.Url
	err := row.Scan(&modelUrl.ID, &modelUrl.Code, &modelUrl.Url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Url{}, nil
		}

		return models.Url{}, fmt.Errorf("%s: %v", op, err)
	}

	return modelUrl, nil
}
