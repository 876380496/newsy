package store

import (
	"context"
)

type StateRepository struct {
	store *SQLiteStore
}

func NewStateRepository(store *SQLiteStore) *StateRepository {
	return &StateRepository{store: store}
}

func (r *StateRepository) SetFetchSuccess(ctx context.Context, sourceKey string) error {
	_, err := r.store.DB.ExecContext(ctx, `
		INSERT INTO source_state (source_key, last_fetched_at, last_error, updated_at)
		VALUES (?, CURRENT_TIMESTAMP, '', CURRENT_TIMESTAMP)
		ON CONFLICT(source_key) DO UPDATE SET
			last_fetched_at = CURRENT_TIMESTAMP,
			last_error = '',
			updated_at = CURRENT_TIMESTAMP
	`, sourceKey)
	return err
}

func (r *StateRepository) SetFetchError(ctx context.Context, sourceKey string, fetchErr error) error {
	_, err := r.store.DB.ExecContext(ctx, `
		INSERT INTO source_state (source_key, last_error, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(source_key) DO UPDATE SET
			last_error = excluded.last_error,
			updated_at = CURRENT_TIMESTAMP
	`, sourceKey, fetchErr.Error())
	return err
}
