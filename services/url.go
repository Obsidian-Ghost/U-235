package services

import (
	"U-235/models"
	"context"
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"time"
)

type UrlServices interface {
	CreateUrlService(originalUrl string, customShortUrl string, expireTime int64, ctx context.Context) (*models.CreateShortUrlRes, error)
}

type ShortUrlService struct {
}

func CreateUrlService(originalUrl string, customShortUrl string, expireTime int64, ctx context.Context) (*models.CreateShortUrlRes, error) {

}
