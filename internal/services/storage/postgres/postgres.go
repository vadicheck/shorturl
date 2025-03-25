package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/vadicheck/shorturl/internal/models"
	"github.com/vadicheck/shorturl/internal/repository"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
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

func (s *Storage) SaveURL(ctx context.Context, code, url, userID string) (int64, error) {
	const op = "storage.postgres.SaveURL"
	const insertURL = "INSERT INTO public.urls (code, url, user_id) VALUES ($1,$2, $3) RETURNING id"

	stmt, err := s.db.Prepare(insertURL)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			slog.Error("prepare sql error", sl.Err(err))
		}
	}()

	var id int64

	err = stmt.QueryRowContext(ctx, code, url, userID).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				mURL, err := s.GetURLByURL(ctx, url)
				if err != nil {
					return 0, err
				}
				if mURL.ID > 0 {
					return 0, &storage.ExistsURLError{
						OriginalURL: url,
						ShortCode:   mURL.Code,
						Err:         err,
					}
				}
			}

			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLOrCodeExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) SaveBatchURL(
	ctx context.Context,
	dto *[]repository.BatchURLDto,
	userID string,
) (*[]repository.BatchURL, error) {
	entities := make([]repository.BatchURL, 0)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("can't begin transaction: %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			slog.Error("transaction rollback error", sl.Err(err))
		}
	}()

	for _, urlDTO := range *dto {
		_, err := s.SaveURL(ctx, urlDTO.ShortCode, urlDTO.OriginalURL, userID)
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
	const selectByCode = "SELECT id, code, url, user_id, is_deleted FROM urls WHERE code=$1"

	row := s.db.QueryRowContext(ctx, selectByCode, code)

	return s.scan(row, "storage.postgres.GetURLByID")
}

func (s *Storage) GetURLByURL(ctx context.Context, url string) (models.URL, error) {
	const selectByURL = "SELECT id, code, url, user_id, is_deleted FROM urls WHERE url=$1"

	row := s.db.QueryRowContext(ctx, selectByURL, url)

	return s.scan(row, "storage.postgres.GetURLByURL")
}

func (s *Storage) GetUserURLs(ctx context.Context, userID string) ([]models.URL, error) {
	const op = "storage.postgres.GetUserURLs"
	const selectByUserID = "SELECT id, code, url, user_id, is_deleted FROM urls WHERE user_id=$1"

	rows, err := s.db.QueryContext(ctx, selectByUserID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs [%s]: %w", op, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("rows close error", sl.Err(err))
		}
	}()

	var urls []models.URL

	for rows.Next() {
		var url models.URL
		if err := rows.Scan(&url.ID, &url.Code, &url.URL, &url.UserID, &url.IsDeleted); err != nil {
			return nil, fmt.Errorf("failed to scan row[%s]: %w", op, err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error encountered during rows iteration [%s]: %w", op, err)
	}

	return urls, nil
}

func (s *Storage) DeleteShortURLs(ctx context.Context, urls []string, userID string) error {
	const op = "storage.postgres.DeleteShortURLs"
	const deleteURLs = "UPDATE public.urls SET is_deleted = true WHERE user_id = $1 AND code = ANY($2)"

	stmt, err := s.db.Prepare(deleteURLs)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			slog.Error("prepare sql error", sl.Err(err))
		}
	}()

	_, err = stmt.ExecContext(ctx, userID, pq.Array(urls))
	if err != nil {
		return fmt.Errorf("can't execute DeleteShortURLs %s: %w", op, err)
	}

	return nil
}

func (s *Storage) scan(row *sql.Row, op string) (models.URL, error) {
	var modelURL models.URL
	err := row.Scan(&modelURL.ID, &modelURL.Code, &modelURL.URL, &modelURL.UserID, &modelURL.IsDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.URL{}, nil
		}

		return models.URL{}, fmt.Errorf("%s: %v", op, err)
	}

	return modelURL, nil
}
