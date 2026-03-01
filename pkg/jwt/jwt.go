package jwt

import (
	"errors"
	"strconv"
	"time"

	jwtsdk "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken          = errors.New("invalid token")
	ErrUnexpectedSigningAlgo = errors.New("unexpected signing method")
)

type Claims struct {
	jwtsdk.RegisteredClaims
	Email string `json:"email"`
}

type Manager struct {
	Secret []byte
	TTL    time.Duration
}

func (m Manager) NewToken(userID int64, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwtsdk.RegisteredClaims{
			Subject:   "user",
			Issuer:    "task-tracker",
			IssuedAt:  jwtsdk.NewNumericDate(now),
			ExpiresAt: jwtsdk.NewNumericDate(now.Add(m.TTL)),
			ID:        strconv.FormatInt(userID, 10),
		},
		Email: email,
	}

	token := jwtsdk.NewWithClaims(jwtsdk.SigningMethodHS256, claims)
	return token.SignedString(m.Secret)
}

type Parser struct {
	Secret []byte
}

func (p Parser) ParseUserID(token string) (int64, error) {
	parsed, err := jwtsdk.ParseWithClaims(token, &Claims{}, func(t *jwtsdk.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtsdk.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSigningAlgo
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
