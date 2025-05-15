package models

import "time"

type CreateShortUrlReq struct {
	OriginalUrl    string `json:"original_url" validate:"required,url"`
	ExpireTime     int64  `json:"expire_time" validate:"required,min=0"`
	CustomShortUrl string `json:"custom_short_url"` //Optional
}

type CreateShortUrlRes struct {
	OriginalUrl string        `json:"original_url" validate:"required,url"`
	ShortUrl    string        `json:"short_url" validate:"required,url"`
	ExpireTime  time.Duration `json:"expire_time" validate:"required,min=0"`
}
