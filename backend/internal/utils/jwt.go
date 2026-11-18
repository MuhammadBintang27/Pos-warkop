package utils

import (
	"fmt"
	"time"

	"warkop/internal/config"
	"warkop/internal/types"

	"github.com/golang-jwt/jwt/v5"
)

type JWTPayload struct {
	UserID  int64      `json:"uid"`
	KlienID int64      `json:"kid"`
	Nama    string     `json:"nama"`
	Role    types.Role `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, klienID int64, nama string, role types.Role) (string, error) {
	cfg := config.Get()
	claims := JWTPayload{
		UserID:  userID,
		KlienID: klienID,
		Nama:    nama,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWTExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
}

func VerifyToken(tokenStr string) (*JWTPayload, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTPayload{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(config.Get().JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	payload, ok := token.Claims.(*JWTPayload)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return payload, nil
}
