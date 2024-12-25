package middleware

import (
	"net/http"
	"strings"
	"encoding/json"
	"Ledger/pkg/auth"
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
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtService.ValidateToken(bearerToken[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		r = r.WithContext(SetUserContext(r.Context(), claims))
		next.ServeHTTP(w, r)
	}
}

func (m *authMiddleware) AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r.Context())
		if claims == nil || !m.jwtService.IsAdmin(claims) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Admin privileges required",
			})
			return
		}
		next.ServeHTTP(w, r)
	}
}