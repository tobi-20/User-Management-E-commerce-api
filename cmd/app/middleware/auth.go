package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

const UserIDKey = "userID" // context key

func CheckJWT(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})

			if err != nil {

				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid claims", http.StatusUnauthorized)
				return
			}

			exp := int64(claims["exp"].(float64))
			if time.Now().Unix() > exp {
				http.Error(w, "token is expired", http.StatusUnauthorized)
				return
			}

			userIDFloat, ok := claims["sub"].(float64)
			if !ok {
				http.Error(w, "user does not exist", http.StatusUnauthorized)
				return
			}

			userID := int64(userIDFloat)

			//extend the initial context amd add the key-value pair UserIDKey: userID into it
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
