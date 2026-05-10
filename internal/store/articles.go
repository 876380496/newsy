package store

import (
	"context"
	"database/sql"
	"fmt"

	"newsy/internal/domain"
)

type ArticleRepository struct {
	store *SQLiteStore
}

func NewArticleRepository(store *SQLiteStore) *ArticleRepository {
	return &ArticleRepository{store: store}
}

func (r *ArticleRepository) UpsertArticles(ctx context.Context, articles []domain.Article) error {
	if len(articles) == 0 {
		return nil
	}

	tx, err := r.store.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO articles (
			source_key, external_id, title, link, author, summary, content, published_at, is_read, is_starred
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source_key, external_id) DO UPDATE SET
			title = excluded.title,
			link = excluded.link,
			author = excluded.author,
			summary = excluded.summary,
			content = excluded.content,
			published_at = excluded.published_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, article := range articles {
		if _, err := stmt.ExecContext(
			ctx,
			article.SourceKey,
			article.ExternalID,
			article.Title,
			article.Link,
			article.Author,
			article.Summary,
			article.Content,
			nullableTime(article.PublishedAt),
			article.IsRead,
			article.IsStarred,
		); err != nil {
			return fmt.Errorf("upsert article %s: %w", article.ExternalID, err)
		}
	}

	return tx.Commit()
}

func (r *ArticleRepository) ListBySource(ctx context.Context, sourceKey string) ([]domain.Article, error) {
	query := `
		SELECT id, source_key, external_id, title, link, author, summary, content, published_at, is_read, is_starred, created_at
		FROM articles
	`
	args := []any{}
	if sourceKey != "" {
		query += " WHERE source_key = ?"
		args = append(args, sourceKey)
	}
	query += " ORDER BY published_at DESC, created_at DESC"

	rows, err := r.store.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []domain.Article
	for rows.Next() {
		var article domain.Article
		var publishedAt sql.NullTime
		if err := rows.Scan(
			&article.ID,
			&article.SourceKey,
			&article.ExternalID,
			&article.Title,
			&article.Link,
			&article.Author,
			&article.Summary,
			&article.Content,
			&publishedAt,
			&article.IsRead,
			&article.IsStarred,
			&article.CreatedAt,
		); err != nil {
			return nil, err
		}
		if publishedAt.Valid {
			article.PublishedAt = publishedAt.Time
		}
		articles = append(articles, article)
	}

	return articles, rows.Err()
}

func (r *ArticleRepository) SetRead(ctx context.Context, id int64, isRead bool) error {
	_, err := r.store.DB.ExecContext(ctx, `UPDATE articles SET is_read = ? WHERE id = ?`, isRead, id)
	return err
}

func (r *ArticleRepository) SetStarred(ctx context.Context, id int64, isStarred bool) error {
	_, err := r.store.DB.ExecContext(ctx, `UPDATE articles SET is_starred = ? WHERE id = ?`, isStarred, id)
	return err
}
