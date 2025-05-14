package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtClaims struct {
	UserId uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}
