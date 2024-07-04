package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	internalauth "github.com/zaytcevcom/msa/internal/server/auth"
	"github.com/zaytcevcom/msa/internal/storage"
)

var signingKey = []byte("demo-secret-key")

type Auth struct {
	logger  Logger
	storage Storage
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Storage interface {
	GetByUsername(ctx context.Context, username string) *storage.PasswordDTO
}

func New(logger Logger, storage Storage) *Auth {
	return &Auth{
		logger:  logger,
		storage: storage,
	}
}

func (a *Auth) Auth(_ context.Context, header http.Header) (int, error) {
	authHeader := header.Get("Authorization")
	if authHeader == "" {
		return 0, errors.New("unauthorized")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, errors.New("unauthorized")
	}

	tokenString := parts[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signature method")
		}
		return signingKey, nil
	})
	if err != nil {
		return 0, errors.New("unauthorized")
	}

	if !token.Valid {
		return 0, errors.New("unauthorized")
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := int(claims["sub"].(float64))

	return userID, nil
}

func (a *Auth) Login(ctx context.Context, username string, password string) (*internalauth.Token, error) {
	passwordData := a.storage.GetByUsername(ctx, username)

	if passwordData == nil {
		return nil, errors.New("invalid login or password")
	}

	hash := sha256.Sum256([]byte(password))
	hashString := hex.EncodeToString(hash[:])

	if passwordData.PasswordHash != hashString {
		return nil, errors.New("invalid login or password")
	}

	userID := passwordData.ID

	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	JWT := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := JWT.SignedString(signingKey)
	if err != nil {
		return nil, errors.New("jwt token failed")
	}

	token := internalauth.Token{
		UserID:      userID,
		AccessToken: tokenString,
	}
	return &token, nil
}
