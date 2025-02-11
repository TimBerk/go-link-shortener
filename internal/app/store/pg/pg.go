package pg

import (
	"context"
	"sync"

	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type PostgresStore struct {
	db  *pgxpool.Pool
	gen store.Generator
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

func NewPgStore(connString string, gen store.Generator) (*PostgresStore, error) {
	ctx := context.Background()

	pgStore, err := NewPgPool(ctx, connString)
	if err != nil {
		return pgStore, err
	}

	pgStore.gen = gen
	pgStore.createTable(ctx)

	return pgStore, nil
}

func (pg *PostgresStore) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *PostgresStore) Close() {
	pg.db.Close()
}

func (pg *PostgresStore) createTable(ctx context.Context) error {
	query := `
    CREATE TABLE IF NOT EXISTS short_urls (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        original_url TEXT NOT NULL,
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

func (pg *PostgresStore) AddURL(originalURL string) (string, error) {
	ctx := context.Background()

	record, err := pg.getRecordByOriginalURL(ctx, originalURL)
	if err == nil && record.OriginalURL != "" && record.ShortURL != "" {
		return record.ShortURL, nil
	} else if err != pgx.ErrNoRows {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"uri": originalURL,
		}).Error("Error checking existing URL")
		return "", err
	}

	shortURL := pg.gen.Next()

	record, err = pg.getRecordByShortURL(ctx, shortURL)
	if err == nil && record.OriginalURL != "" && record.ShortURL != "" {
		return pg.AddURL(originalURL)
	} else if err != pgx.ErrNoRows {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"uri": shortURL,
		}).Error("Error checking existing short URL")
		return "", err
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

func (pg *PostgresStore) GetOriginalURL(shortURL string) (string, bool) {
	ctx := context.Background()

	record, err := pg.getRecordByShortURL(ctx, shortURL)
	if err == nil {
		logrus.Error(err)
		return record.OriginalURL, true
	}
	return "", false
}
