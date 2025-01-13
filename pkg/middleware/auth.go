package middleware

import (
	"Ledger/pkg/auth"
	"Ledger/pkg/response"
	"fmt"
	"net/http"
	"strings"
)

type AuthMiddleware interface {
	Authenticate(next http.HandlerFunc) http.HandlerFunc
	AdminOnly(next http.HandlerFunc) http.HandlerFunc
}

type authMiddleware struct {
	jwtService auth.JWTService
}

func NewAuthMiddleware(jwtService auth.JWTService) AuthMiddleware {
	return &authMiddleware{
		jwtService: jwtService,
	}
}

func (m *authMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.WriteError(w, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		fmt.Printf("Auth header: %s\n", authHeader)

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			response.WriteError(w, http.StatusUnauthorized, "Invalid token format. Must be 'Bearer <token>'")
			return
		}

		fmt.Printf("Token: %s\n", bearerToken[1])

		claims, err := m.jwtService.ValidateToken(bearerToken[1])
		if err != nil {
			response.WriteError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid token: %v", err))
			return
		}

		fmt.Printf("Claims: %+v\n", claims)

		r = r.WithContext(SetUserInContext(r.Context(), claims))
		next.ServeHTTP(w, r)
	}
}

func (m *authMiddleware) AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r.Context())
		if claims == nil || !m.jwtService.IsAdmin(claims) {
			response.WriteError(w, http.StatusForbidden, "Admin privileges required")
			return
		}
		next.ServeHTTP(w, r)
	}
}
