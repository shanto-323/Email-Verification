package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"email-auth/internal/model"
	middleware "email-auth/middleware"
	pkg "email-auth/pkg"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/ksuid"
)

type Server struct {
	ipAddr string
	cache  pkg.Cache
}

func NewServer(ipAddr string, client *redis.Client) *Server {
	return &Server{
		ipAddr: ipAddr,
		cache:  pkg.NewRedisCatch(client),
	}
}

func (s *Server) Run() error {
	router := http.NewServeMux()
	router.HandleFunc("/signup", middleware.Ratelimit(middleware.HandleFunc(s.signUpHandler), s.cache))
	router.HandleFunc("/signin", middleware.Ratelimit(middleware.HandleFunc(s.signInHandler), s.cache))
	router.HandleFunc("/verify", middleware.HandleFunc(s.emailVerificationHandler))
	router.HandleFunc("/newcode", middleware.Ratelimit(middleware.HandleFunc(s.newEmailVerificationHandler), s.cache))

	log.Println("gateway running on port ", s.ipAddr, " ......")
	return http.ListenAndServe(s.ipAddr, router)
}

func (s *Server) signUpHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("invalid method")
	}

	client := &model.ClientModel{}
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		return err
	}
	client.ID = ksuid.New().String()
	client.Purpose = model.SIGNUP
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
		model.EmailVerificationData{
			UID:   client.ID,
			Email: client.Email,
			EXP:   1 * time.Minute,
		},
	); err != nil {
		return err
	}

	return pkg.WriteJson(w, http.StatusOK, model.UserIdentifier{UID: client.ID})
}

func (s *Server) signInHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("invalid method")
	}

	client := &model.ClientModel{}
	if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
		return err
	}

	// CHECK IF USER EXISTS -> IN DB.
	// FOR NOW IMAGINE THIS ID COMING FROM DATABASE.
	client.ID = ksuid.New().String()
	client.Purpose = model.SIGNIN
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
		model.EmailVerificationData{
			UID:   client.ID,
			Email: client.Email,
			EXP:   1 * time.Minute,
		},
	); err != nil {
		return err
	}

	return pkg.WriteJson(w, http.StatusOK, model.UserIdentifier{UID: client.ID})
}

func (s *Server) emailVerificationHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("invalid method")
	}

	fa2 := &model.TwoFARequest{}
	if err := json.NewDecoder(r.Body).Decode(&fa2); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client_model := &model.ClientModel{}
	if err := s.cache.GetValue(ctx, fa2.UID, client_model); err != nil {
		return err
	}

	verify_model := &model.EmailVerificationData{}
	if err := s.cache.GetValue(ctx, fa2.Code, verify_model); err != nil {
		return err
	}

	if client_model.Email != verify_model.Email {
		return fmt.Errorf("invalid code")
	}

	if client_model.Purpose == model.SIGNIN {
		return pkg.WriteJson(w, http.StatusCreated, "sign up success")
	}
	return pkg.WriteJson(w, http.StatusCreated, "user created")
}

func (s *Server) newEmailVerificationHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("invalid method")
	}

	id := &model.UserIdentifier{}
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &model.ClientModel{}
	if err := s.cache.GetValue(ctx, id.UID, client); err != nil {
		return err
	}

	if err := s.sendEmailVerification(
		ctx,
		model.EmailVerificationData{
			UID:   client.ID,
			Email: client.Email,
			EXP:   1 * time.Minute,
		},
	); err != nil {
		return err
	}

	return pkg.WriteJson(w, http.StatusCreated, "email sent")
}

func (s *Server) sendEmailVerification(ctx context.Context, verifyModel model.EmailVerificationData) error {
	to := verifyModel.Email
	subject := "varification"
	body := pkg.GenerateSixDigitCode()

	value, err := json.Marshal(verifyModel)
	if err != nil {
		return err
	}

	if err := s.cache.SetValue(ctx, body, value, 1*time.Minute); err != nil {
		return nil
	}

	if err := pkg.SendEmailVerification(to, subject, body); err != nil {
		return err
	}

	return nil
}
