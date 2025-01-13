package auth

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTService interface {
	GenerateToken(userID uint, email string, isAdmin bool) (string, error)
	ValidateToken(tokenString string) (*JWTClaim, error)
	IsAdmin(claims *JWTClaim) bool
}

type jwtService struct {
	secretKey string
	issuer    string
}

type JWTClaim struct {
	UserID  uint   `json:"user_id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func NewJWTService() JWTService {
	return &jwtService{
		secretKey: os.Getenv("JWT_SECRET_KEY"),
		issuer:    "ledger.app",
	}
}

func (j *jwtService) GenerateToken(userID uint, email string, isAdmin bool) (string, error) {
	expirationHours, _ := strconv.Atoi(os.Getenv("JWT_EXPIRATION_HOURS"))
	claims := &JWTClaim{
		UserID:  userID,
		Email:   email,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expirationHours))),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *jwtService) ValidateToken(tokenString string) (*JWTClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		return nil, fmt.Errorf("couldn't parse claims")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

func (j *jwtService) IsAdmin(claims *JWTClaim) bool {
	return claims.IsAdmin
}
