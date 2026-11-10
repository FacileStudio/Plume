package middleware

import "net/http"

// NoStore returns middleware that sets response headers preventing caching
// and disabling referrer leakage and content-type sniffing.
func NoStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Cache-Control", "no-store, max-age=0")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, request)
	})
}
