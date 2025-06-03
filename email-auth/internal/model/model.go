package model

import "time"

const (
	SIGNUP = "signup"
	SIGNIN = "signin"
)

type ClientModel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Purpose  string `json:"purpose"`
}

type EmailVerificationData struct {
	UID   string        `json:"id"`
	Email string        `json:"email"`
	EXP   time.Duration `json:"exp"`
}

type UserIdentifier struct {
	UID string `json:"id"`
}

type TwoFARequest struct {
	UID  string `json:"id"`
	Code string `json:"fa2"` // 2FA Code
}
