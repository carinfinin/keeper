package jwtr

import (
	"errors"
	"fmt"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func Generate(user *models.User, tokenType string, cfg *config.Config) (string, error) {

	// Определяем длительность по типу токена
	var duration time.Duration
	switch tokenType {
	case "access":
		duration = cfg.AccessTokenDuration
	case "refresh":
		duration = cfg.RefreshTokenDuration
	default:
		return "", errors.New("invalid token type")
	}

	//// Загрузка ключа
	//privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.PrivateKey))
	//if err != nil {
	//	return "", fmt.Errorf("failed to parse private key: %w", err)
	//}

	// Создание claims
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":                user.ID,
		"preferred_username": user.Login,
		"exp":                now.Add(duration).Unix(),
		"iat":                now.Unix(),
		"nbf":                now.Unix(),
		//"jti":                uuid.New().String(),
		"iss":   cfg.Addr,
		"aud":   cfg.JWTAudience,
		"type":  tokenType,
		"scope": "openid profile email",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Заголовок для защиты от downgrade-атак
	token.Header["alg"] = "RS256"
	token.Header["kid"] = cfg.JWTKeyID

	return token.SignedString(cfg.PrivateKey)
}

func Decode(token string, cfg *config.Config) (int64, error) {
	claims := jwt.MapClaims{}
	tp, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return cfg.PublicKey, nil
	})
	if err != nil {
		return 0, err
	}
	if !tp.Valid {
		return 0, fmt.Errorf("is not valid")
	}

	fmt.Println(claims)

	uid, ok := claims["uid"].(float64)
	if !ok {
		return 0, fmt.Errorf("not converce in int64 from %v", claims)
	}
	//jwt.RegisteredClaim
	return int64(uid), nil
}
