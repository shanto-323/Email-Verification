package main

import "time"

type SignUpModel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VerifyModel struct {
	UID   string        `json:"uid"`
	Email string        `json:"email"`
	EXP   time.Duration `json:"exp"`
}
