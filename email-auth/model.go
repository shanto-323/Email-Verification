package main

import "time"

type SignUpModel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VerifyModel struct {
	UID   string        `json:"id"`
	Email string        `json:"email"`
	EXP   time.Duration `json:"exp"`
}

type IDModel struct {
	UID string `json:"id"`
}

type FA2Model struct {
	UID string `json:"id"`
	FA2 string `json:"fa2"` // 2FA Code
}
