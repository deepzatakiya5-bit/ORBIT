package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const defaultTokenTTL = 30 * 24 * time.Hour // 30 days

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string, ttl time.Duration) (*TokenService, error) {
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if ttl <= 0 {
		ttl = defaultTokenTTL
	}
	return &TokenService{secret: []byte(secret), ttl: ttl}, nil
}

type claims struct {
	jwt.RegisteredClaims
}

func (s *TokenService) Issue(userID uuid.UUID) (string, error) {
	now := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			Issuer:    "orbit",
		},
	})
	return t.SignedString(s.secret)
}

func (s *TokenService) Parse(token string) (uuid.UUID, error) {
	parsed, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	c, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid || c.Subject == "" {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	userID, err := uuid.Parse(c.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid subject")
	}
	return userID, nil
}
