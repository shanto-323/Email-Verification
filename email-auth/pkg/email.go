package pkg

import (
	"gopkg.in/gomail.v2"
)

func SendEmailVerification(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@shanto")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	dial := gomail.NewDialer("mailhog", 1025, "", "")
	return dial.DialAndSend(m)
}
