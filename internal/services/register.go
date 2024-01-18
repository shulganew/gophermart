package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// User registration service
type Register struct {
	stor Registrar
}

type Registrar interface {
	Register(ctx context.Context, user model.User) error
	GetByLogin(ctx context.Context, login string) (*model.User, error)
}

func NewRegister(stor Registrar) *Register {
	return &Register{stor: stor}
}

// Register new user in market
func (r *Register) NewUser(ctx context.Context, user model.User) (existed bool, err error) {
	// Generate UUID for new user.
	uuid, err := uuid.NewV7()
	if err != nil {
		zap.S().Errorln("UUID error generation")
		return true, err
	}

	user.UUID = &uuid

	// Set hash as user password.
	hash, err := r.HashPassword(user.Password)
	if err != nil {
		zap.S().Errorln("Error creating hash from password")
		return true, err
	}
	user.Password = hash

	// Add user to database.
	err = r.stor.Register(ctx, user)
	if err != nil {
		var pgErr *pgconn.PgError
		// If URL exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			zap.S().Infoln("User exist: ", user)
			return true, err
		}
	}

	return false, nil
}

// Validate user in market, if sucsess it return user's id.
func (r *Register) IsValid(ctx context.Context, untrusted *model.User) (userID *uuid.UUID, isValid bool) {

	// Get User from storage
	user, err := r.stor.GetByLogin(ctx, untrusted.Login)
	zap.S().Infoln("User form db: ", user)
	if err != nil {
		zap.S().Errorln("User not found by login")
		return nil, false
	}

	// Check pass is correct
	err = r.CheckPassword(untrusted.Password, user.Password)
	if err != nil {
		zap.S().Errorln("Pass not valid: ", err)
		return nil, false
	}

	return user.UUID, true
}

// HashPassword returns the bcrypt hash of the password
func (r Register) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword checks if the provided password is correct or not
func (r Register) CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Claims for JWT token
type Claims struct {
	jwt.RegisteredClaims
	UserID *uuid.UUID
}

// Create JWT token
func BuildJWTString(user *model.User, pass string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.TokenExp)),
		},
		UserID: user.UUID,
	})

	tokenString, err := token.SignedString([]byte(pass))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Retrive user's UUID from JWT string
func GetUserIDJWT(tokenString string, pass string) (userID *uuid.UUID, err error) {
	claims := &Claims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(pass), nil
	})

	return claims.UserID, err
}

// Create jwt token from string
func GetJWT(tokenString string, pass string) (token *jwt.Token, err error) {
	claims := &Claims{}
	token, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(pass), nil
	})

	return token, err
}

// Check JWT is Set to Header
func GetHeaderJWT(header http.Header) (jwt string, isSet bool) {
	jwt = header.Get("Authorization")[len("Bearer "):]
	if jwt == "" {
		return "", false
	}
	return jwt, true

}