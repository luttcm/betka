package notifications

import (
	"fmt"
	"log"
	"net/smtp"

	"bet/backend/internal/config"
)

type Message struct {
	To      string
	Subject string
	Body    string
}

type Sender interface {
	Send(message Message) error
}

func NewSenderFromConfig(cfg config.Config) Sender {
	if cfg.SMTPHost == "" || cfg.SMTPPort == "" {
		return &logSender{from: cfg.EmailFrom}
	}

	return &smtpSender{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		username: cfg.SMTPUsername,
		password: cfg.SMTPPassword,
		from:     cfg.EmailFrom,
	}
}

type logSender struct {
	from string
}

func (s *logSender) Send(message Message) error {
	log.Printf("email stub: from=%s to=%s subject=%q body=%q", s.from, message.To, message.Subject, message.Body)
	return nil
}

type smtpSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func (s *smtpSender) Send(message Message) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	headers := ""
	headers += fmt.Sprintf("From: %s\r\n", s.from)
	headers += fmt.Sprintf("To: %s\r\n", message.To)
	headers += fmt.Sprintf("Subject: %s\r\n", message.Subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += "Content-Type: text/plain; charset=UTF-8\r\n\r\n"

	payload := []byte(headers + message.Body)

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	if err := smtp.SendMail(addr, auth, s.from, []string{message.To}, payload); err != nil {
		return fmt.Errorf("send smtp email: %w", err)
	}

	return nil
}
