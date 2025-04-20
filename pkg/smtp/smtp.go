package smtp

import (
	"fmt"
	smtpPkg "net/smtp"
	"os"
)

type ItfSmtp interface {
	CreateSmtp(userEmail string, otp string) error
}

type smtp struct {
	auth smtpPkg.Auth
	mail string
}

func New() ItfSmtp {
	mail := os.Getenv("SMTP_MAIL")
	password := os.Getenv("SMTP_PASSWORD")
	auth := smtpPkg.PlainAuth("", mail, password, "smtp.gmail.com")

	return &smtp{auth: auth, mail: mail}
}

func (s *smtp) CreateSmtp(userEmail string, otp string) error {
	to := []string{userEmail}

	message := []byte(fmt.Sprintf("To: %s\r\nSubject: Your OTP\r\n\r\nHello %s, this is from %s, this is your OTP: %s",
		userEmail, userEmail, s.mail, otp))

	msgArray := message

	err := smtpPkg.SendMail("smtp.gmail.com:587", s.auth, s.mail, to, msgArray)
	if err != nil {
		return err
	}

	return nil
}
