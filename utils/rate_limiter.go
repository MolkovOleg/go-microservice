package utils

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)


func NewFixedRateLimiter() func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(1000, 5000)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				next.ServeHTTP(w, r)
				return 
			}
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := limiter.Wait(ctx); err != nil {
				http.Error(w, "Too many requsets", http.StatusTooManyRequests)
				return 
			}
			next.ServeHTTP(w, r)
		})
	}
}