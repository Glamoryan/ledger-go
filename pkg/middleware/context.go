package middleware

import (
	"Ledger/pkg/auth"
	"context"
)

type contextKey string

const UserContextKey contextKey = "user"

func SetUserInContext(ctx context.Context, claims *auth.JWTClaim) context.Context {
	return context.WithValue(ctx, UserContextKey, claims)
}

func GetUserFromContext(ctx context.Context) *auth.JWTClaim {
	claims, ok := ctx.Value(UserContextKey).(*auth.JWTClaim)
	if !ok {
		return nil
	}
	return claims
}
