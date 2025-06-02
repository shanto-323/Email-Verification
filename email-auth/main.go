package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/ksuid"
	"github.com/tinrab/retry"
)

type Server struct {
	ipAddr string
	cache  Cache
}

func NewServer(ipAddr string, client *redis.Client) *Server {
	return &Server{
		ipAddr: ipAddr,
		cache:  NewRedisCatch(client),
	}
}

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

	server := NewServer(cfg.IpAddr, redisClient)

	router := http.NewServeMux()
	router.HandleFunc("/signup", handleFunc(server.signUpHandler))
	// router.HandleFunc("/verify", handleFunc(server.sendEmailVarificationHandler))

	log.Fatal(http.ListenAndServe(server.ipAddr, router))
}

func (s *Server) signUpHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("invalid method")
	}

	client := &SignUpModel{}
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		return err
	}
	client.ID = ksuid.New().String()
	log.Println(client)

	value, err := json.Marshal(client)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.cache.SetValue(ctx, client.ID, value, 20*time.Minute); err != nil {
		return nil
	}

	if err := s.sendEmailVarification(
		ctx,
		VerifyModel{
			UID:   client.ID,
			Email: client.Email,
			EXP:   1 * time.Minute,
		},
	); err != nil {
		return err
	}

	return WriteJson(w, http.StatusCreated, "code sent")
}

func (s *Server) sendEmailVarification(ctx context.Context, verifyModel VerifyModel) error {
	to := verifyModel.Email
	subject := "varification"
	body := GenerateSixDigitCode()

	value, err := json.Marshal(verifyModel)
	if err != nil {
		return err
	}

	if err := s.cache.SetValue(ctx, body, value, 1*time.Minute); err != nil {
		return nil
	}

	if err := sendEmailVarification(to, subject, body); err != nil {
		return err
	}

	return nil
}

func handleFunc(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJson(w, http.StatusBadRequest, err)
			return
		}
	}
}

func WriteJson(w http.ResponseWriter, status int, msg any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(msg)
}

func GenerateSixDigitCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := 100000 + r.Intn(900000)
	return strconv.Itoa(code)
}
