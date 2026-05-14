package middleware

import (
	"net/http"
	"strings"
)

var corsAllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
var corsAllowedHeaders = []string{"Accept", "Authorization", "Content-Type"}

func CORS(domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			origin := request.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, request)
				return
			}

			if !isAllowedOrigin(origin, domain) {
				next.ServeHTTP(w, request)
				return
			}

			header := w.Header()
			header.Add("Vary", "Origin")
			header.Add("Vary", "Access-Control-Request-Method")
			header.Add("Vary", "Access-Control-Request-Headers")
			header.Set("Access-Control-Allow-Origin", origin)
			header.Set("Access-Control-Allow-Methods", strings.Join(corsAllowedMethods, ", "))
			header.Set("Access-Control-Allow-Headers", strings.Join(corsAllowedHeaders, ", "))
			header.Set("Access-Control-Max-Age", "600")

			if request.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, request)
		})
	}
}

func isAllowedOrigin(origin string, domain string) bool {
	return domain == "*" || origin == domain
}
