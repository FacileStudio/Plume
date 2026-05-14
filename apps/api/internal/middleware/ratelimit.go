package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*bucket
	capacity   float64
	refillRate float64
	ttl        time.Duration
}

func NewRateLimiter(requestsPerMinute int, burst int) *RateLimiter {
	if requestsPerMinute < 1 {
		requestsPerMinute = 1
	}
	if burst < 1 {
		burst = requestsPerMinute
	}
	limiter := &RateLimiter{
		buckets:    make(map[string]*bucket),
		capacity:   float64(burst),
		refillRate: float64(requestsPerMinute) / 60.0,
		ttl:        10 * time.Minute,
	}
	go limiter.sweep()
	return limiter
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	current, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.capacity - 1, lastRefill: now}
		return true
	}
	elapsed := now.Sub(current.lastRefill).Seconds()
	current.tokens += elapsed * rl.refillRate
	if current.tokens > rl.capacity {
		current.tokens = rl.capacity
	}
	current.lastRefill = now
	if current.tokens < 1 {
		return false
	}
	current.tokens--
	return true
}

func (rl *RateLimiter) sweep() {
	ticker := time.NewTicker(rl.ttl)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.ttl)
		for key, b := range rl.buckets {
			if b.lastRefill.Before(cutoff) {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			if !rl.allow(clientKey(request)) {
				w.Header().Set("Retry-After", "60")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    "resource_exhausted",
						"message": "too many requests",
					},
				})
				return
			}
			next.ServeHTTP(w, request)
		})
	}
}

func clientKey(request *http.Request) string {
	ip := request.RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	return ip
}

func NoStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Cache-Control", "no-store, max-age=0")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, request)
	})
}
