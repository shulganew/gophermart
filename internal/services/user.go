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
	"github.com/lib/pq"
	"github.com/shulganew/gophermart/internal/config"
	"github.com/shulganew/gophermart/internal/entities"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// User creation, registration, validation and autentification service.
type UserService struct {
	stor UserRepo
}

type UserRepo interface {
	AddUser(ctx context.Context, login string, hash string) (*uuid.UUID, error)
	GetByLogin(ctx context.Context, login string) (*entities.User, error)
}

func NewUserService(stor UserRepo) *UserService {
	return &UserService{stor: stor}
}

// Register new user in market.
func (r *UserService) CreateUser(ctx context.Context, login string, password string) (userID *uuid.UUID, existed bool, err error) {
	// Set hash as user password.
	hash, err := r.HashPassword(password)
	if err != nil {
		zap.S().Errorln("Error creating hash from password")
		return nil, true, err
	}

	// Add user to database.
	userID, err = r.stor.AddUser(ctx, login, hash)
	if err != nil {
		var pgErr *pq.Error
		// If URL exist in DataBase
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			zap.S().Infoln("User with login alredy existed: ", login)
			return nil, true, nil
		}
		return nil, false, err
	}

	return userID, false, nil
}

// Validate user in market, if sucsess it return user's id.
func (r *UserService) IsValid(ctx context.Context, login string, pass string) (userID *uuid.UUID, isValid bool) {
	// Get User from storage
	user, err := r.stor.GetByLogin(ctx, login)
	zap.S().Infof("User form db: %v \n", user)
	if err != nil {
		zap.S().Infoln("User not found by login. ", err)
		return nil, false
	}

	// Check pass is correct
	err = r.CheckPassword(pass, user.PassHash)
	if err != nil {
		zap.S().Errorln("Pass not valid: ", err)
		return nil, false
	}

	return &user.UUID, true
}

// HashPassword returns the bcrypt hash of the password.
func (r UserService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword checks if the provided password is correct or not.
func (r UserService) CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Create JWT token.
func BuildJWTString(userID uuid.UUID, pass string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, entities.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(pass))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Retrive user's UUID from JWT string.
func GetUserIDJWT(tokenString string, pass string) (userID uuid.UUID, err error) {
	claims := &entities.Claims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(pass), nil
	})

	return claims.UserID, err
}

// Create jwt token from string.
func GetJWT(tokenString string, pass string) (token *jwt.Token, err error) {
	claims := &entities.Claims{}
	token, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(pass), nil
	})

	return token, err
}

// Check JWT is Set to Header.
func GetHeaderJWT(header http.Header) (jwt string, isSet bool) {
	auth := header.Get("Authorization")
	if auth == "" {
		return "", false
	}

	return auth, true
}
