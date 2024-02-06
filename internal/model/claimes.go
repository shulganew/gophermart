package model

import (
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
)

// Claims for JWT token.
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}
