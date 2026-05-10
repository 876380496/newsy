package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"newsy/internal/logging"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	DB *sql.DB
}

func Open(path string) (*SQLiteStore, error) {
	logging.Infof("sqlite open begin path=%s", path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		logging.Errorf("sqlite open failed path=%s err=%v", path, err)
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	logging.Infof("sqlite connection limits set max_open=1 max_idle=1")

	pragmas := []struct {
		name  string
		query string
	}{
		{name: "journal_mode", query: "PRAGMA journal_mode = WAL"},
		{name: "synchronous", query: "PRAGMA synchronous = NORMAL"},
		{name: "busy_timeout", query: "PRAGMA busy_timeout = 5000"},
	}
	for _, pragma := range pragmas {
		if _, err := db.ExecContext(context.Background(), pragma.query); err != nil {
			logging.Errorf("sqlite pragma set failed path=%s pragma=%s err=%v", path, pragma.name, err)
			_ = db.Close()
			return nil, err
		}
		logging.Infof("sqlite pragma set path=%s pragma=%s", path, pragma.name)
	}

	store := &SQLiteStore{DB: db}
	if err := store.migrate(context.Background()); err != nil {
		logging.Errorf("sqlite migrate failed path=%s err=%v", path, err)
		_ = db.Close()
		return nil, err
	}

	logging.Infof("sqlite open complete path=%s", path)
	return store, nil
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}
	logging.Infof("sqlite close")
	return s.DB.Close()
}

func (s *SQLiteStore) migrate(ctx context.Context) error {
	logging.Infof("sqlite migrate begin")
	statements := []string{
		`CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_key TEXT NOT NULL,
			external_id TEXT NOT NULL,
			title TEXT NOT NULL,
			link TEXT NOT NULL,
			author TEXT NOT NULL DEFAULT '',
			summary TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL DEFAULT '',
			published_at TIMESTAMP,
			is_read INTEGER NOT NULL DEFAULT 0,
			is_starred INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(source_key, external_id)
		);`,
		`CREATE TABLE IF NOT EXISTS source_state (
			source_key TEXT PRIMARY KEY,
			last_fetched_at TIMESTAMP,
			last_error TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, stmt := range statements {
		if _, err := s.DB.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate sqlite: %w", err)
		}
	}

	logging.Infof("sqlite migrate complete")
	return nil
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
