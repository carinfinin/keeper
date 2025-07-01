package jwtr

import (
	"errors"
	"fmt"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// Generate генерирует jwt.
func Generate(user *models.User, tokenType string, cfg *config.Config) (string, error) {

	var duration time.Duration
	switch tokenType {
	case "access":
		duration = time.Duration(cfg.AccessTokenDuration) * time.Minute
	case "refresh":
		duration = time.Duration(cfg.RefreshTokenDuration) * time.Hour
	default:
		return "", errors.New("invalid token type")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"uid":                user.ID,
		"did":                user.DeviceID,
		"preferred_username": user.Login,
		"exp":                now.Add(duration).Unix(),
		"iat":                now.Unix(),
		"iss":                cfg.Addr,
		"aud":                cfg.JWTAudience,
		"type":               tokenType,
		"scope":              "openid profile email",
		"alg":                "RS256",
		"kid":                cfg.JWTKeyID,
		//"jti":                uuid.New().String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(cfg.PrivateKey)
}

// JwtData структура для дальнейшей предачи в контекст.
type JwtData struct {
	UserID   float64
	DeviceID float64
}

// Decode.
func Decode(token string, cfg *config.Config) (*JwtData, error) {

	claims := jwt.MapClaims{}
	tp, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return cfg.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !tp.Valid {
		return nil, fmt.Errorf("is not valid")
	}

	uid, ok := claims["uid"].(float64)
	if !ok {
		return nil, fmt.Errorf("not converce in int64 from %v", claims)
	}
	did, ok := claims["did"].(float64)
	if !ok {
		return nil, fmt.Errorf("not converce in int64 from %v", claims)
	}
	return &JwtData{
		UserID:   uid,
		DeviceID: did,
	}, nil
}
