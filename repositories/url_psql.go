package repositories

import (
	"U-235/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"log"
)

type UrlsPsql interface {
	SaveUrl(ctx context.Context, UrlInfo *models.ShortenedUrlInfoReq) (*models.ShortenedUrlInfoRes, *models.PsqlRollback, error)
	GetUrlInfoByUserIdAndShortUrl(ctx context.Context, userId uuid.UUID, shortUrl string) (*models.ShortenedUrlInfoRes, error)
	GetUrlInfoByUserIdAndUrlRecordId(ctx context.Context, userId uuid.UUID, urlRecordId uuid.UUID) (*models.ShortenedUrlInfoRes, error)
	DeleteUrlRecord(ctx context.Context, UserId uuid.UUID, UrlRecordId uuid.UUID) error
	SoftDeleteUrl(ctx context.Context, userId uuid.UUID, urlId uuid.UUID) error
	GetUserUrls(ctx context.Context, userID uuid.UUID, offset, limit int, isActive *bool) ([]models.ShortenedUrlInfoRes, error)
	CountUserUrls(ctx context.Context, userID uuid.UUID, isActive *bool) (int64, error)
	SetUrlState(ctx context.Context, userId uuid.UUID, urlId uuid.UUID, isActive bool) error
	UrlRecordExists(ctx context.Context, urlID uuid.UUID) (bool, error)
	ExtendExpiry(ctx context.Context, userId uuid.UUID, urlId uuid.UUID, hours int) error
	MarkUrlAsExpired(ctx context.Context, shortUrl string) error
}

type UrlsPsqlImpl struct {
	db     *sql.DB
	gormDB *gorm.DB
}

func NewUrlsPsql(db *sql.DB, gormDB *gorm.DB) UrlsPsql {
	return &UrlsPsqlImpl{
		db:     db,
		gormDB: gormDB,
	}
}

// DeleteUrlRecord - Used only for rollback purposes when creating a new short URL, in case the operation fails.
func (u *UrlsPsqlImpl) DeleteUrlRecord(ctx context.Context, UserId uuid.UUID, UrlRecordId uuid.UUID) error {
	query := `DELETE FROM shortened_urls WHERE user_id = $1 AND id = $2`
	_, err := u.db.ExecContext(ctx, query, UserId, UrlRecordId)
	return err
}

func (u *UrlsPsqlImpl) SoftDeleteUrl(ctx context.Context, userId uuid.UUID, urlId uuid.UUID) error {
	query := `
       UPDATE shortened_urls
       SET is_active = false, expires_at = NOW()
       WHERE user_id = $1 AND id = $2
    `
	result, err := u.db.ExecContext(ctx, query, userId, urlId)
	if err != nil {
		return fmt.Errorf("failed to soft delete URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no URL found to delete")
	}

	return nil
}

func (u *UrlsPsqlImpl) SaveUrl(ctx context.Context, urlInfo *models.ShortenedUrlInfoReq) (*models.ShortenedUrlInfoRes, *models.PsqlRollback, error) {
	// Start a transaction
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	//Wrap error handling in closure - GoLand IDE suggestion
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {

		}
	}(tx)

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

func (u *UrlsPsqlImpl) GetUserUrls(ctx context.Context, userID uuid.UUID, offset, limit int, isActive *bool) ([]models.ShortenedUrlInfoRes, error) {
	var urls []models.ShortenedUrlInfoRes

	query := u.gormDB.WithContext(ctx).
		Table("shortened_urls").
		Where("user_id = ?", userID)

	// Apply active status filter if provided
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Apply pagination and order by creation date (newest first)
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&urls).Error

	if err != nil {
		return nil, err
	}

	return urls, nil
}

func (u *UrlsPsqlImpl) CountUserUrls(ctx context.Context, userID uuid.UUID, isActive *bool) (int64, error) {
	var count int64

	query := u.gormDB.WithContext(ctx).
		Table("shortened_urls").
		Where("user_id = ?", userID)

	// Apply active status filter if provided
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (u *UrlsPsqlImpl) GetUrlInfoByUserIdAndUrlRecordId(ctx context.Context, userId uuid.UUID, urlRecordId uuid.UUID) (*models.ShortenedUrlInfoRes, error) {
	query := `
		SELECT id, user_id, original_url, short_url, expires_at, is_active, created_at
		FROM shortened_urls
		WHERE user_id = $1 AND id = $2;
	`

	var urlInfo models.ShortenedUrlInfoRes

	err := u.db.QueryRowContext(ctx, query, userId, urlRecordId).Scan(
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
			return nil, fmt.Errorf("no URL found for user ID %s and short URL %s: %w", userId.String(), urlRecordId, err)
		}
		// Log or wrap unexpected DB error
		return nil, fmt.Errorf("failed to fetch URL info from DB: %w", err)
	}

	return &urlInfo, nil
}

func (u *UrlsPsqlImpl) SetUrlState(ctx context.Context, userId uuid.UUID, urlId uuid.UUID, isActive bool) error {
	query := `
		UPDATE shortened_urls
		SET is_active = $1
		WHERE user_id = $2 AND id = $3
	`
	_, err := u.db.ExecContext(ctx, query, isActive, userId, urlId)
	if err != nil {
		return fmt.Errorf("failed to update URL state: %w", err)
	}
	return nil
}

func (u *UrlsPsqlImpl) UrlRecordExists(ctx context.Context, urlID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM shortened_urls WHERE id = $1)`

	err := u.db.QueryRowContext(ctx, query, urlID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (u *UrlsPsqlImpl) ExtendExpiry(ctx context.Context, userId uuid.UUID, urlId uuid.UUID, hours int) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `UPDATE shortened_urls SET expires_at = expires_at + make_interval(hours => $1) WHERE id = $2 AND user_id = $3`

	result, err := tx.ExecContext(ctx, query, hours, urlId, userId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

func (u *UrlsPsqlImpl) MarkUrlAsExpired(ctx context.Context, shortUrl string) error {
	query := `
        UPDATE shortened_urls 
        SET is_active = false, 
            expires_at = NOW() 
        WHERE short_url = $1 AND is_active = true
    `

	result, err := u.db.ExecContext(ctx, query, shortUrl)
	if err != nil {
		return fmt.Errorf("failed to mark URL as expired: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Printf("No active URL found for shortUrl: %s", shortUrl)
	}

	return nil
}
