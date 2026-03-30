package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"transport-payments/models"
)

var jwtSecret = []byte("WOOFWOOFBARKBARKLOVELOVEILOVEYOUPUPPY")

func GenerateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"login":    user.Login,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func ExtractClaimsFromRequest(r *http.Request) (jwt.MapClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return nil, fmt.Errorf("invalid authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

	return ParseToken(tokenString)
}
