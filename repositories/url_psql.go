package repositories

import (
	"U-235/models"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type UrlsPsql interface {
	SaveUrl(ctx context.Context, UrlInfo *models.ShortenedUrlInfoReq) (*models.ShortenedUrlInfoRes, *models.PsqlRollback, error)
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

func (u *UrlsPsqlImpl) SaveUrl(ctx context.Context, UrlInfo *models.ShortenedUrlInfoReq) (*models.ShortenedUrlInfoRes, *models.PsqlRollback, error) {
	return nil, nil, nil
}
