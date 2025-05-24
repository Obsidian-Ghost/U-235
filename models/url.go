package models

import (
	"github.com/google/uuid"
	"time"
)

type ShortenedUrlInfoRes struct {
	Id          uuid.UUID `json:"id"`
	UserId      uuid.UUID `json:"user_id"`
	OriginalUrl string    `json:"original_url" validate:"required,url"`
	ShortUrl    string    `json:"short_url" validate:"required,url"`
	ExpiresAt   time.Time `json:"expires_at" validate:"required,min=0"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type ShortenedUrlInfoReq struct {
	UserId      uuid.UUID `json:"user_id"`
	OriginalUrl string    `json:"original_url" validate:"required,url"`
	ShortUrl    string    `json:"short_url" validate:"required,url"`
	ExpiresAt   time.Time `json:"expires_at" validate:"required,min=0"`
	IsActive    bool      `json:"is_active"`
}

type CreateShortUrlReq struct {
	OriginalUrl    string `json:"original_url" validate:"required,url"`
	ExpireTime     int64  `json:"expire_time" validate:"required"`
	CustomShortUrl string `json:"custom_short_url"` //Optional
}

type DeleteShortUrlReq struct {
	UserId      uuid.UUID `json:"user_id"`
	UrlRecordId uuid.UUID `json:"url_record_id"`
}

type PsqlRollback struct {
	UserId      uuid.UUID `json:"user_id"`
	UrlRecordId uuid.UUID `json:"url_record_id"`
}

type ExtendExpiry struct {
	UrlId uuid.UUID `json:"url_id"`
	Hours int       `json:"hours" validate:"required,lt=73"`
}
