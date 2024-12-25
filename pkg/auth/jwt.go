package auth

import (
    "errors"
    "time"
    "github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("your-secret-key")

type Claims struct {
    UserID uint
    Email  string
    Role   string
    jwt.RegisteredClaims
}

type JWTService interface {
    GenerateToken(userID uint, email, role string) (string, error)
    ValidateToken(tokenStr string) (*Claims, error)
    IsAdmin(claims *Claims) bool
}

type jwtService struct{}

func NewJWTService() JWTService {
    return &jwtService{}
}

func (s *jwtService) GenerateToken(userID uint, email, role string) (string, error) {
    claims := &Claims{
        UserID: userID,
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}

func (s *jwtService) ValidateToken(tokenStr string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })

    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, errors.New("invalid token")
    }

    return claims, nil
}

func (s *jwtService) IsAdmin(claims *Claims) bool {
    return claims.Role == "admin"
} 