package auth

import (
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserId string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

var secretKey []byte

const secretDefaultValue = "your_jwt_secret"

func init() {
	secretKey = []byte(env.GetString("JWT_SECRET", secretDefaultValue))
}

func GenerateToken(userId string, roles []string, expiresAt time.Time) (string, error) {
	claims := UserClaims{
		UserId: userId,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
