package usecase

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	Secret []byte
	TTL    time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

func (m JWTManager) NewToken(userID int64, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user",
			Issuer:    "task-tracker",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.TTL)),
			ID:        "",
		},
		Email: email,
	}
	claims.RegisteredClaims.ID = fmtInt64(userID)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.Secret)
}

func fmtInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}
