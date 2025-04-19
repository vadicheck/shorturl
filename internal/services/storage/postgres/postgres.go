// Package postgres provides an implementation of the Storage interface using PostgreSQL.
// It supports saving, retrieving, and deleting URL data, as well as batch processing and user-specific URL management.
// The package also includes migration handling to ensure the database schema is up-to-date.
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

// Storage is the PostgreSQL implementation of the Storage interface for managing URL data.
// It provides methods for saving, retrieving, and deleting URLs, as well as handling batch operations.
type Storage struct {
	// db is the underlying PostgreSQL database connection.
	db *sql.DB
}

// New initializes and returns a new Storage instance.
// It connects to the PostgreSQL database using the provided database DSN (Data Source Name),
// and applies any necessary migrations to ensure the database schema is up-to-date.
// It returns the Storage instance or an error if the connection or migration fails.
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

// PingContext checks the database connection by executing a ping operation.
// It returns an error if the connection cannot be established.
func (s *Storage) PingContext(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// SaveURL saves a URL with a unique code to the database.
// If the URL already exists in the database, an error is returned.
// It returns the ID of the newly saved URL or an error if the save operation fails.
func (s *Storage) SaveURL(ctx context.Context, code, url, userID string) (int64, error) {
	const op = "storage.postgres.SaveURL"
	const insertURL = "INSERT INTO public.urls (code, url, user_id) VALUES ($1,$2, $3) RETURNING id"

	stmt, err := s.db.Prepare(insertURL)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if errStmtClose := stmt.Close(); errStmtClose != nil {
			slog.Error("prepare sql error", sl.Err(errStmtClose))
		}
	}()

	var id int64
	err = stmt.QueryRowContext(ctx, code, url, userID).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				mURL, errGetURL := s.GetURLByURL(ctx, url)
				if errGetURL != nil {
					return 0, errGetURL
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

// SaveBatchURL saves multiple URLs in a batch operation to the database.
// It processes a list of BatchURLDto, saving each URL individually. If a URL already exists, it is skipped,
// and the batch continues with the next URL.
// It returns a list of BatchURL instances with the correlation ID and short code for each URL.
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
		if errRollback := tx.Rollback(); errRollback != nil {
			slog.Error("transaction rollback error", sl.Err(errRollback))
		}
	}()

	for _, urlDTO := range *dto {
		_, errSaveURL := s.SaveURL(ctx, urlDTO.ShortCode, urlDTO.OriginalURL, userID)
		if errSaveURL != nil {
			if errors.Is(errSaveURL, storage.ErrURLOrCodeExists) {
				mURL, errGetURL := s.GetURLByURL(ctx, urlDTO.OriginalURL)
				if errGetURL != nil {
					return nil, fmt.Errorf("failed to retrieve URL: %w", errGetURL)
				}

				entities = append(entities, repository.BatchURL{
					CorrelationID: urlDTO.CorrelationID,
					ShortCode:     mURL.Code,
				})
				continue
			} else {
				return nil, fmt.Errorf("failed to save URL: %w", errSaveURL)
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

// GetURLByID retrieves a URL from the database using its code.
// It returns the URL corresponding to the provided code or an error if no matching URL is found.
func (s *Storage) GetURLByID(ctx context.Context, code string) (models.URL, error) {
	const selectByCode = "SELECT id, code, url, user_id, is_deleted FROM urls WHERE code=$1"

	row := s.db.QueryRowContext(ctx, selectByCode, code)

	return s.scan(row, "storage.postgres.GetURLByID")
}

// GetURLByURL retrieves a URL from the database using the full URL.
// It returns the URL corresponding to the provided URL or an error if no matching URL is found.
func (s *Storage) GetURLByURL(ctx context.Context, url string) (models.URL, error) {
	const selectByURL = "SELECT id, code, url, user_id, is_deleted FROM urls WHERE url=$1"

	row := s.db.QueryRowContext(ctx, selectByURL, url)

	return s.scan(row, "storage.postgres.GetURLByURL")
}

// GetUserURLs retrieves all URLs associated with a user from the database.
// It returns a list of URLs for the specified user or an error if no URLs are found.
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

// DeleteShortURLs marks URLs as deleted for a given user in the database.
// It updates the `is_deleted` field to true for each of the provided short URLs.
func (s *Storage) DeleteShortURLs(ctx context.Context, urls []string, userID string) error {
	const op = "storage.postgres.DeleteShortURLs"
	const deleteURLs = "UPDATE public.urls SET is_deleted = true WHERE user_id = $1 AND code = ANY($2)"

	stmt, err := s.db.Prepare(deleteURLs)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if errStmtClose := stmt.Close(); errStmtClose != nil {
			slog.Error("prepare sql error", sl.Err(errStmtClose))
		}
	}()

	_, err = stmt.ExecContext(ctx, userID, pq.Array(urls))
	if err != nil {
		return fmt.Errorf("can't execute DeleteShortURLs %s: %w", op, err)
	}

	return nil
}

// scan is a helper function that scans a row from the database and maps it to a models.URL.
// It returns the scanned URL or an error if the scan operation fails.
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
