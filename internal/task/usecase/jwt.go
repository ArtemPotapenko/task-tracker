package usecase

import (
	"errors"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

var errUnexpectedSigningMethod = errors.New("unexpected signing method")

type JWTParser struct {
	Secret []byte
}

type Claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

func (p JWTParser) ParseUserID(token string) (int64, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errUnexpectedSigningMethod
		}
		return p.Secret, nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return 0, ErrInvalidToken
	}
	if claims.Subject != "user" || claims.Issuer != "task-tracker" {
		return 0, ErrInvalidToken
	}

	if claims.ID == "" {
		return 0, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(claims.ID, 10, 64)
	if err != nil {
		return 0, ErrInvalidToken
	}
	return userID, nil
}
