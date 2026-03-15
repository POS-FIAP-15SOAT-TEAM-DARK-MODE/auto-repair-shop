package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserId string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateToken(secretKey, userId string, roles []string, expiresAt time.Time) (string, error) {
	claims := UserClaims{
		UserId: userId,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
