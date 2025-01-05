package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"

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

func (s *Storage) SaveURL(ctx context.Context, code, url string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO main.urls (code, url) VALUES (?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, code, url)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLOrCodeExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURLByID(ctx context.Context, code string) (models.URL, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, code, url FROM urls WHERE code=?", code)

	return s.scan(row, "storage.sqlite.GetURLByID")
}

func (s *Storage) GetURLByURL(ctx context.Context, url string) (models.URL, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, code, url FROM urls WHERE url=?", url)

	return s.scan(row, "storage.sqlite.GetURLByURL")
}

func (s *Storage) scan(row *sql.Row, op string) (models.URL, error) {
	var modelURL models.URL
	err := row.Scan(&modelURL.ID, &modelURL.Code, &modelURL.URL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.URL{}, nil
		}

		return models.URL{}, fmt.Errorf("%s: %v", op, err)
	}

	return modelURL, nil
}
