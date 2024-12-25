package middleware

import (
    "context"
    "Ledger/pkg/auth"
)

type contextKey string

const userContextKey contextKey = "user"

func SetUserContext(ctx context.Context, claims *auth.Claims) context.Context {
    return context.WithValue(ctx, userContextKey, claims)
}

func GetUserFromContext(ctx context.Context) *auth.Claims {
    if claims, ok := ctx.Value(userContextKey).(*auth.Claims); ok {
        return claims
    }
    return nil
} 