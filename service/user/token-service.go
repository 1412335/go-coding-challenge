package user

import (
	"fmt"

	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/utils"
	"go.uber.org/zap"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	jwt.StandardClaims
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type TokenService struct {
	logger     log.Factory
	jwtManager *utils.JWTManager
}

func NewTokenService(config *configs.JWT) *TokenService {
	return &TokenService{
		logger:     log.With(zap.String("srv", "token")),
		jwtManager: utils.NewJWTManager(config),
	}
}

func (t *TokenService) Generate(user *User) (string, error) {
	claims := Claims{
		StandardClaims: t.jwtManager.GetStandardClaims(),
		ID:             user.ID,
		Email:          user.Email,
	}
	return t.jwtManager.Generate(claims)
}

func (t *TokenService) Verify(accessToken string) (*Claims, error) {
	claims, err := t.jwtManager.Verify(accessToken, &Claims{})
	if err != nil {
		t.logger.Bg().Error("verify token failed", zap.Error(err))
		return nil, err
	}
	uc, ok := claims.(*Claims)
	if !ok {
		t.logger.Bg().Error("invalid", zap.Any("claims", claims))
		return nil, fmt.Errorf("invalid token claims")
	}
	return uc, nil
}
