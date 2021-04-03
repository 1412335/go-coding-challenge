package utils

import (
	"fmt"
	"time"

	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"

	"github.com/dgrijalva/jwt-go"
)

type JWTManager struct {
	issuer        string
	secretKey     []byte
	tokenDuration time.Duration
}

func NewJWTManager(config *configs.JWT) *JWTManager {
	return &JWTManager{
		issuer:        config.Issuer,
		secretKey:     []byte(config.SecretKey),
		tokenDuration: config.Duration,
	}
}

func (manager *JWTManager) GetStandardClaims() jwt.StandardClaims {
	return jwt.StandardClaims{
		ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		Issuer:    manager.issuer,
	}
}

func (manager *JWTManager) Generate(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(manager.secretKey)
}

func (manager *JWTManager) Verify(accessToken string, claims jwt.Claims) (jwt.Claims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return manager.secretKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
