package middleware

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func makeToken(t *testing.T, secret string, claims jwt.MapClaims) string {
    t.Helper()

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString([]byte(secret))
    if err != nil {
        t.Fatalf("failed to sign token: %v", err)
    }
    return signed
}

func TestJWTAuth(t *testing.T) {
    validClaims := jwt.MapClaims{
        "user_id": float64(123),
        "role":    "Админ",
        "exp":     time.Now().Add(1 * time.Hour).Unix(),
    }

    validToken := makeToken(t, testSecret, validClaims)
    wrongSecretToken := makeToken(t, "wrong-secret", validClaims)

    tests := []struct {
        name       string
        authHeader string
        wantStatus int
    }{
        {
            name:       "missing authorization header",
            authHeader: "",
            wantStatus: http.StatusUnauthorized,
        },
        {
            name:       "invalid authorization header format",
            authHeader: "Token abc",
            wantStatus: http.StatusUnauthorized,
        },
        {
            name:       "invalid signature",
            authHeader: "Bearer " + wrongSecretToken,
            wantStatus: http.StatusUnauthorized,
        },
        {
            name: "missing user_id claim",
            authHeader: "Bearer " + makeToken(t, testSecret, jwt.MapClaims{
                "role": "Админ",
                "exp":  time.Now().Add(1 * time.Hour).Unix(),
            }),
            wantStatus: http.StatusUnauthorized,
        },
        {
            name: "missing role claim",
            authHeader: "Bearer " + makeToken(t, testSecret, jwt.MapClaims{
                "user_id": float64(123),
                "exp":     time.Now().Add(1 * time.Hour).Unix(),
            }),
            wantStatus: http.StatusUnauthorized,
        },
        {
            name:       "valid token",
            authHeader: "Bearer " + validToken,
            wantStatus: http.StatusOK,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mw := JWTAuth(testSecret)

            next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                if strings.Contains(tt.name, "valid token") {
                    userID, ok := r.Context().Value(CtxUserID).(int32)
                    if !ok || userID != 123 {
                        t.Fatalf("expected user_id=123 in context, got %v (ok=%v)", userID, ok)
                    }

                    role, ok := r.Context().Value(CtxRole).(string)
                    if !ok || role != "Админ" {
                        t.Fatalf("expected role=Админ in context, got %v (ok=%v)", role, ok)
                    }
                }

                w.WriteHeader(http.StatusOK)
            })

            req := httptest.NewRequest(http.MethodGet, "/protected", nil)
            if tt.authHeader != "" {
                req.Header.Set("Authorization", tt.authHeader)
            }
            rr := httptest.NewRecorder()

            mw(next).ServeHTTP(rr, req)

            if rr.Code != tt.wantStatus {
                t.Fatalf("want status %d, got %d, body=%s", tt.wantStatus, rr.Code, rr.Body.String())
            }
        })
    }
}