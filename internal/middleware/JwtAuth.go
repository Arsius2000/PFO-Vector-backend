package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
    CtxUserID ctxKey = "user_id"
    CtxRole   ctxKey = "role"
)

func JWTAuth(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "missing authorization header", http.StatusUnauthorized)
                return
            }

            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            if tokenString == authHeader {
                http.Error(w, "invalid authorization header", http.StatusUnauthorized)
                return
            }

            token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
                return []byte(secret), nil
            })
            if err != nil || !token.Valid {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            claims, ok := token.Claims.(jwt.MapClaims)
            if !ok {
                http.Error(w, "invalid token claims", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), CtxUserID, claims["user_id"])
            ctx = context.WithValue(ctx, CtxRole, claims["role"])

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}