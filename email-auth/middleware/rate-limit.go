package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	pkg "email-auth/pkg"

	"github.com/redis/go-redis/v9"
)

func Ratelimit(next http.HandlerFunc, cache pkg.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if _, err := cache.ValueExists(ctx, ip); err != redis.Nil {
			pkg.NewMessage(w, http.StatusTooManyRequests, "limit exceeds")
			return
		}

		if err := cache.SetValue(ctx, ip, nil, time.Minute); err != nil {
			pkg.NewMessage(w, http.StatusTooManyRequests, "error in setting value")
			return
		}

		next.ServeHTTP(w, r)
	}
}
