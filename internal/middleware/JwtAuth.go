package middleware

import (
	"context"
	"errors"
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
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, errors.New("invalid signing method")
                }
                if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
                    return nil, errors.New("unexpected signing method")
                }
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
			

			userIDf, ok := claims["user_id"].(float64)
			if !ok {
				http.Error(w, "invalid user_id claim", http.StatusUnauthorized)
				return
			}

			role, ok := claims["role"].(string)
			if !ok {
				http.Error(w, "invalid role claim", http.StatusUnauthorized)
				return
			}

            ctx := context.WithValue(r.Context(), CtxUserID, int32(userIDf))
            ctx = context.WithValue(ctx, CtxRole, role)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}