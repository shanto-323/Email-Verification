package main

import (
	"context"
	"fmt"
	"log"
	"time"

	controller "email-auth/internal/controller"

	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"github.com/tinrab/retry"
)

type Config struct {
	RedisUrl string `envconfig:"REDIS_URL"`
	IpAddr   string `envconfig:"IP_ADDR"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	var redisClient *redis.Client
	retry.ForeverSleep(
		2*time.Second,
		func(_ int) error {
			opts, err := redis.ParseURL(cfg.RedisUrl)
			if err != nil {
				return err
			}
			redisClient = redis.NewClient(opts)
			pong, err := redisClient.Ping(context.Background()).Result()
			if err != nil {
				return err
			}
			if pong != "PONG" {
				return fmt.Errorf("unexpected ping result")
			}
			return nil
		},
	)

	server := controller.NewServer(fmt.Sprintf(":%s", cfg.IpAddr), redisClient)
	log.Fatal(server.Run())
}
