package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgconn"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/repository"
	"github.com/vadicheck/shorturl/internal/services/storage"
)

type Storage struct {
	db *sql.DB
}

func New(databaseDsn string) (*Storage, error) {
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	m, err := migrate.New("file://migrations", databaseDsn)
	if err != nil {
		log.Panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No migrations needed")
		} else {
			log.Panic(err)
		}
	}

	return &Storage{db: db}, nil
}

func (s *Storage) PingContext(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Storage) SaveURL(ctx context.Context, code, url string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO public.urls (code, url) VALUES ($1,$2) RETURNING id")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var id int64

	err = stmt.QueryRowContext(ctx, code, url).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLOrCodeExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) SaveBatchURL(ctx context.Context, dto *[]repository.BatchURLDto) (*[]repository.BatchURL, error) {
	entities := make([]repository.BatchURL, 0)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("can't begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, urlDTO := range *dto {
		_, err := s.SaveURL(ctx, urlDTO.ShortCode, urlDTO.OriginalURL)
		if err != nil {
			if errors.Is(err, storage.ErrURLOrCodeExists) {
				mURL, err := s.GetURLByURL(ctx, urlDTO.OriginalURL)
				if err != nil {
					return nil, fmt.Errorf("failed to retrieve URL: %w", err)
				}

				entities = append(entities, repository.BatchURL{
					CorrelationID: urlDTO.CorrelationID,
					ShortCode:     mURL.Code,
				})
				continue
			} else {
				return nil, fmt.Errorf("failed to save URL: %w", err)
			}
		}

		entities = append(entities, repository.BatchURL{
			CorrelationID: urlDTO.CorrelationID,
			ShortCode:     urlDTO.ShortCode,
		})
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("can't commit transaction: %w", err)
	}

	return &entities, nil
}

func (s *Storage) GetURLByID(ctx context.Context, code string) (models.URL, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, code, url FROM urls WHERE code=$1", code)

	return s.scan(row, "storage.postgres.GetURLByID")
}

func (s *Storage) GetURLByURL(ctx context.Context, url string) (models.URL, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, code, url FROM urls WHERE url=$1", url)

	return s.scan(row, "storage.postgres.GetURLByURL")
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
