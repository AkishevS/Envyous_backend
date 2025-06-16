package middleware

import (
	"net/http"
)

var APIKey string

func APIKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || apiKey != APIKey {
			http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})

}
