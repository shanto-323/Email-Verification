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
	router.HandleFunc("/signup", handleFunc(server.SignUpHandler))
	router.HandleFunc("/verify", handleFunc(server.EmailVerificationHandler))
	router.HandleFunc("/newcode", handleFunc(server.NewEmailVerificationHandler))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", server.ipAddr), router))
}

func (s *Server) SignUpHandler(w http.ResponseWriter, r *http.Request) error {
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

	if err := s.sendEmailVerification(
		ctx,
		VerifyModel{
			UID:   client.ID,
			Email: client.Email,
			EXP:   1 * time.Minute,
		},
	); err != nil {
		return err
	}

	return WriteJson(w, http.StatusOK, IDModel{UID: client.ID})
}

func (s *Server) EmailVerificationHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("invalid method")
	}

	fa2 := &FA2Model{}
	if err := json.NewDecoder(r.Body).Decode(&fa2); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client_model := &SignUpModel{}
	if err := s.cache.GetValue(ctx, fa2.UID, client_model); err != nil {
		return err
	}
	log.Println(fmt.Sprintln("client_model", client_model))

	verify_model := &VerifyModel{}
	if err := s.cache.GetValue(ctx, fa2.FA2, verify_model); err != nil {
		return err
	}
	log.Println(fmt.Sprintln("verify_model", verify_model)) // Found = issued fa2 (2fa) key

	if client_model.Email != verify_model.Email {
		return fmt.Errorf("invalid code")
	}

	return WriteJson(w, http.StatusCreated, "user created")
}

func (s *Server) NewEmailVerificationHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("invalid method")
	}

	id := &IDModel{}
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &SignUpModel{}
	if err := s.cache.GetValue(ctx, id.UID, client); err != nil {
		return err
	}
	log.Println(fmt.Sprintln("client_model", client))

	if err := s.sendEmailVerification(
		ctx,
		VerifyModel{
			UID:   client.ID,
			Email: client.Email,
			EXP:   1 * time.Minute,
		},
	); err != nil {
		return err
	}

	return WriteJson(w, http.StatusCreated, "email sent")
}

func (s *Server) sendEmailVerification(ctx context.Context, verifyModel VerifyModel) error {
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

	if err := sendEmailVerification(to, subject, body); err != nil {
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
