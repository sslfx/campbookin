package api_gateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

func (gw *Gateway) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Println(">>> MIDDLEWARE HIT")
		fmt.Println("AUTH:", req.Header.Get("Authorization"))
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "miss token", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return gw.secret, nil
		})
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, "token expired or invalid", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}
		userID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "user_id not found in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), "userID", userID)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
func (gw *Gateway) withAuthContext(req *http.Request) context.Context {
	userID := req.Context().Value("userID").(string)

	md := metadata.New(map[string]string{
		"user_id": userID,
	})
	return metadata.NewOutgoingContext(req.Context(), md)
}
