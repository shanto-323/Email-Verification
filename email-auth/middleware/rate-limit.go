package middleware

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	in "email-auth/internal"

	"github.com/redis/go-redis/v9"
)

func Ratelimit(next http.HandlerFunc, cache in.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if _, err := cache.ValueExists(ctx, ip); err != redis.Nil {
			log.Println("limit exceeds")
			return
		}

		if err := cache.SetValue(ctx, ip, nil, time.Minute); err != nil {
			log.Println("error in setting value")
			return
		}

		log.Println("cache set for ip", ip)
		next.ServeHTTP(w, r)
	}
}
