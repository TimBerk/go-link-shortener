package pg

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	models "github.com/TimBerk/go-link-shortener/internal/app/models/batch"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type PostgresStore struct {
	db  *pgxpool.Pool
	gen store.Generator
	cfg *config.Config
}

type PgRecord struct {
	ID          string
	OriginalURL string
	ShortURL    string
}

func NewPgPool(ctx context.Context, connString string) (*PostgresStore, error) {
	var pgInstance *PostgresStore
	var pgOnce sync.Once
	var pgErr error

	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, connString)
		if err != nil {
			pgErr = err
			return
		}

		pgInstance = &PostgresStore{db: db}
	})

	if pgErr != nil {
		logrus.WithFields(logrus.Fields{
			"err": pgErr,
		}).Error("unable to create connection pool")
		return nil, pgErr
	}

	return pgInstance, nil
}

func NewPgStore(gen store.Generator, cfg *config.Config) (*PostgresStore, error) {
	ctx := context.Background()

	pgStore, err := NewPgPool(ctx, cfg.DatabaseDSN)
	if err != nil {
		return pgStore, err
	}

	pgStore.gen = gen
	pgStore.cfg = cfg
	pgStore.createTable(ctx)

	return pgStore, nil
}

func (pg *PostgresStore) Ping(ctx context.Context) error {
	connection, err := pgx.Connect(ctx, pg.cfg.DatabaseDSN)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("unable to create connection")
		return err
	}
	defer connection.Close(ctx)

	err = connection.Ping(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("failed to ping database")
	}
	return err
}

func (pg *PostgresStore) Close() {
	pg.db.Close()
}

func (pg *PostgresStore) createTable(ctx context.Context) error {
	query := `
    CREATE TABLE IF NOT EXISTS short_urls (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        original_url TEXT NOT NULL UNIQUE,
        short_url VARCHAR(6) NOT NULL UNIQUE
    );`
	_, err := pg.db.Exec(ctx, query)
	return err
}

func (pg *PostgresStore) getRecordByOriginalURL(ctx context.Context, originalURL string) (PgRecord, error) {
	var record PgRecord
	query := `SELECT * FROM short_urls WHERE original_url = $1`
	err := pg.db.QueryRow(ctx, query, originalURL).Scan(&record.ID, &record.OriginalURL, &record.ShortURL)
	return record, err
}

func (pg *PostgresStore) getRecordByShortURL(ctx context.Context, shortURL string) (PgRecord, error) {
	var record PgRecord
	query := `SELECT * FROM short_urls WHERE short_url = $1`
	err := pg.db.QueryRow(ctx, query, shortURL).Scan(&record.ID, &record.OriginalURL, &record.ShortURL)
	return record, err
}

func (pg *PostgresStore) insertRecord(ctx context.Context, originalURL, shortURL string) error {
	query := `INSERT INTO short_urls (original_url, short_url) VALUES ($1, $2)`
	_, err := pg.db.Exec(ctx, query, originalURL, shortURL)
	return err
}

func (pg *PostgresStore) AddURL(ctx context.Context, originalURL string) (string, error) {
	record, err := pg.getRecordByOriginalURL(ctx, originalURL)
	logrus.WithFields(logrus.Fields{
		"originalURL": originalURL,
		"record":      record,
		"err":         err,
	}).Info("Attempt check short URL link")

	if err == nil {
		return record.ShortURL, store.ErrLinkExist
	} else if !errors.Is(err, pgx.ErrNoRows) {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"uri": originalURL,
		}).Error("Error checking existing URL")
		return "", err
	}

	var shortURL string
	for {
		shortURL = pg.gen.Next()
		_, err := pg.getRecordByShortURL(ctx, shortURL)
		logrus.WithFields(logrus.Fields{
			"shortURL": shortURL,
			"err":      err,
		}).Info("Attempt generate new short URL link")

		if errors.Is(err, pgx.ErrNoRows) {
			break
		} else if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
				"uri": shortURL,
			}).Error("failed to check existing short URL")
			return "", err
		}
	}

	if err := pg.insertRecord(ctx, originalURL, shortURL); err != nil {
		logrus.WithFields(logrus.Fields{
			"err":      err,
			"uri":      originalURL,
			"shortUri": shortURL,
		}).Error("Error inserting new URL")
		return "", err
	}

	return shortURL, nil
}

func (pg *PostgresStore) AddURLs(ctx context.Context, urls models.BatchRequest) (models.BatchResponse, error) {
	var responses models.BatchResponse

	tx, err := pg.db.Begin(ctx)
	if err != nil {
		logrus.WithField("err", err).Error("Error starting transaction")
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO short_urls (ID, original_url, short_url) VALUES ($1, $2, $3) ON CONFLICT (short_url) DO NOTHING`
	stmt, err := tx.Prepare(ctx, "insert-tx-stmt", query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	for _, req := range urls {
		record, errRecord := pg.getRecordByOriginalURL(ctx, req.OriginalURL)
		if errRecord == nil {
			responses = append(responses, models.ItemResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      fmt.Sprintf("http://%s/%s", pg.cfg.ServerAddress, record.ShortURL),
			})
			continue
		} else if !errors.Is(errRecord, pgx.ErrNoRows) {
			logrus.WithFields(logrus.Fields{
				"err":         errRecord,
				"originalURL": req.OriginalURL,
			}).Error("failed to check existing original URL")
		}

		var shortURL string
		for {
			shortURL = pg.gen.Next()
			_, err := pg.getRecordByShortURL(ctx, shortURL)
			if errors.Is(err, pgx.ErrNoRows) {
				break
			} else if err != nil {
				logrus.WithFields(logrus.Fields{
					"err": err,
					"uri": shortURL,
				}).Error("failed to check existing short URL")
			}
		}

		_, err := tx.Exec(ctx, stmt.SQL, req.CorrelationID, req.OriginalURL, shortURL)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":      err,
				"ID":       req.CorrelationID,
				"uri":      req.OriginalURL,
				"shortUri": shortURL,
			}).Error("Error inserting URL")
		} else {
			responses = append(responses, models.ItemResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      fmt.Sprintf("http://%s/%s", pg.cfg.ServerAddress, shortURL),
			})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		logrus.WithField("err", err).Error("Error committing transaction")
		return nil, err
	}

	return responses, nil
}

func (pg *PostgresStore) GetOriginalURL(ctx context.Context, shortURL string) (string, bool) {
	record, err := pg.getRecordByShortURL(ctx, shortURL)
	if err == nil {
		return record.OriginalURL, true
	}

	logrus.WithFields(logrus.Fields{
		"uri": shortURL,
		"err": err,
	}).Error("Short URL not found")
	return "", false
}
