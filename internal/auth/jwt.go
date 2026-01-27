package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

type contextKey string

const UserIDKey contextKey = "user_id"

var tokenExp = time.Hour * 24
var secretKey = []byte("my-secret-key-change-in-production")

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		var userID string
		var hasValidCookie bool

		if err == nil {
			token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secretKey, nil
			})

			if err == nil && token.Valid {
				hasValidCookie = true
				if claims, ok := token.Claims.(*Claims); ok {
					userID = claims.UserID
				}
			}
		}

		if userID == "" {
			// we should guarantee here userID is always set in valid cookie
			// otherwise - http 401 and don't run next.ServeHTTP
			if hasValidCookie {
				http.Error(w, "invalid user ID in token", http.StatusUnauthorized)
				return
			}
			userID = uuid.New().String()

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
				UserID: userID,
			})

			tokenString, err := token.SignedString(secretKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    tokenString,
				Path:     "/",
				Expires:  time.Now().Add(tokenExp),
				HttpOnly: true,
				Secure:   false,
			})
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
