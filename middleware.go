package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func makeToken(userID int, secret []byte) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

type ctxKey string

const userIDKey ctxKey = "userID"

func (app *App) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || tokenStr == "" {
			writeError(w, http.StatusUnauthorized, "missing or malformed token", "")
			return
		}

		userID, err := parseToken(tokenStr, app.jwtSecret)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token", "")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

func parseToken(tokenStr string, secret []byte) (int, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{},
		func(t *jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{"HS256"}), // принимаем только свой алгоритм
	)
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}
	return strconv.Atoi(claims.Subject)
}
