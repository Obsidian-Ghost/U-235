package repositories

import (
	"U-235/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type UrlsPsql interface {
	SaveUrl(ctx context.Context, UrlInfo *models.ShortenedUrlInfoReq) (*models.ShortenedUrlInfoRes, *models.PsqlRollback, error)
	GetUrlInfoByUserIdAndShortUrl(ctx context.Context, userId uuid.UUID, shortUrl string) (*models.ShortenedUrlInfoRes, error)
	DeleteUrlRecord(ctx context.Context, UserId uuid.UUID, UrlRecordId uuid.UUID) error
}

type UrlsPsqlImpl struct {
	db *sql.DB
}

func NewUrlsPsql(db *sql.DB) UrlsPsql {
	return &UrlsPsqlImpl{db: db}
}

func (u *UrlsPsqlImpl) DeleteUrlRecord(ctx context.Context, UserId uuid.UUID, UrlRecordId uuid.UUID) error {
	query := `DELETE FROM shortened_urls WHERE user_id = $1 AND id = $2`
	_, err := u.db.ExecContext(ctx, query, UserId, UrlRecordId)
	return err
}

func (u *UrlsPsqlImpl) SaveUrl(ctx context.Context, urlInfo *models.ShortenedUrlInfoReq) (*models.ShortenedUrlInfoRes, *models.PsqlRollback, error) {
	// Start a transaction
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
        INSERT INTO shortened_urls (
            user_id, original_url, short_url, expires_at, is_active
        ) VALUES (
            $1, $2, $3, $4, $5
        ) RETURNING id, user_id, original_url, short_url, expires_at, is_active, created_at;
    `

	var response models.ShortenedUrlInfoRes

	err = tx.QueryRowContext(
		ctx,
		query,
		urlInfo.UserId,
		urlInfo.OriginalUrl,
		urlInfo.ShortUrl,
		urlInfo.ExpiresAt,
		urlInfo.IsActive,
	).Scan(
		&response.Id,
		&response.UserId,
		&response.OriginalUrl,
		&response.ShortUrl,
		&response.ExpiresAt,
		&response.IsActive,
		&response.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, nil, fmt.Errorf("short URL already exists: %w", err)
		}
		return nil, nil, fmt.Errorf("failed to insert shortened url: %w", err)
	}

	rollback := &models.PsqlRollback{
		UserId:      response.UserId,
		UrlRecordId: response.Id,
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &response, rollback, nil
}

func (u *UrlsPsqlImpl) GetUrlInfoByUserIdAndShortUrl(ctx context.Context, userId uuid.UUID, shortUrl string) (*models.ShortenedUrlInfoRes, error) {
	query := `
		SELECT id, user_id, original_url, short_url, expires_at, is_active, created_at
		FROM shortened_urls
		WHERE user_id = $1 AND short_url = $2;
	`

	var urlInfo models.ShortenedUrlInfoRes

	err := u.db.QueryRowContext(ctx, query, userId, shortUrl).Scan(
		&urlInfo.Id,
		&urlInfo.UserId,
		&urlInfo.OriginalUrl,
		&urlInfo.ShortUrl,
		&urlInfo.ExpiresAt,
		&urlInfo.IsActive,
		&urlInfo.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Not found — return a typed error for clarity
			return nil, fmt.Errorf("no URL found for user ID %s and short URL %s: %w", userId.String(), shortUrl, err)
		}
		// Log or wrap unexpected DB error
		return nil, fmt.Errorf("failed to fetch URL info from DB: %w", err)
	}

	return &urlInfo, nil
}
