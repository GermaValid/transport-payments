package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"transport-payments/auth"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ExtractClaimsFromRequest(r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"error":"invalid or missing token"}`)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

func GetClaims(r *http.Request) (jwt.MapClaims, bool) {
	claims, ok := r.Context().Value(ClaimsKey).(jwt.MapClaims)
	return claims, ok
}

func WriteUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"error":"%s"}`, message)
}

func WriteForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprintf(w, `{"error":"%s"}`, message)
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r)
		if !ok {
			WriteUnauthorized(w, "missing claims in context")
			return
		}

		isAdminFloat, ok := claims["is_admin"].(float64)
		if !ok {
			WriteUnauthorized(w, "invalid token payload")
			return
		}

		if int(isAdminFloat) != 1 {
			WriteForbidden(w, "admin access required")
			return
		}

		next(w, r)
	}
}

func GetUserID(r *http.Request) (int64, bool) {
	claims, ok := GetClaims(r)
	if !ok {
		return 0, false
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, false
	}

	return int64(userIDFloat), true
}

func IsAdmin(r *http.Request) bool {
	claims, ok := GetClaims(r)
	if !ok {
		return false
	}

	isAdminFloat, ok := claims["is_admin"].(float64)
	if !ok {
		return false
	}

	return int(isAdminFloat) == 1
}
