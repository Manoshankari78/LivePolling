package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	Secret []byte
	Expiry time.Duration
}

func NewJWTManager(secret string, expiry time.Duration) *JWTManager {
	return &JWTManager{Secret: []byte(secret), Expiry: expiry}
}

func (m *JWTManager) Generate(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.Expiry)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.Secret)
}

func (m *JWTManager) Parse(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.Secret, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return "", errors.New("invalid token claims")
	}
	return claims.Subject, nil
}
