package main

import (
	"context"
	"github.com/go-chi/cors"
	"golang-observer-project/token"
	"net/http"
	"strings"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func CorsMiddleware() func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:3006"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type",
			"X-CSRF-Token", "X-Requested-With", "X-Auth-Token", "X-Auth-User", "X-Auth-Password"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

func authMiddleware(tokenMaker token.Maker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Authorization header
			authHeader := r.Header.Get(authorizationHeaderKey)
			// Check if Authorization header exists
			if len(authHeader) == 0 {
				http.Error(w, "authorization header is not provided", http.StatusUnauthorized)
				return
			}

			// Validate token format
			// Bearer <token>
			fields := strings.Fields(authHeader)
			if len(fields) != 2 {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			// Validate token prefix
			tokenType := strings.ToLower(fields[0])
			if tokenType != authorizationTypeBearer {
				http.Error(w, "invalid authorization type", http.StatusUnauthorized)
				return
			}

			accessToken := fields[1]
			payload, err := tokenMaker.VerifyToken(accessToken)
			if err != nil {
				http.Error(w, "invalid authorization token", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, authorizationPayloadKey, payload)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
