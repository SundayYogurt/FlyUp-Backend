package services

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/smtp"
	"net/url"
	"strings"
	"time"
)

type MailService struct {
	gmailUser     string
	gmailAppPass  string
	mailFrom      string
	mailFromName  string
	subject       string
	verifyBaseURL string
}

func NewMailService(
	gmailUser string,
	gmailAppPass string,
	mailFrom string,
	mailFromName string,
	subject string,
	verifyBaseURL string,
) *MailService {
	return &MailService{
		gmailUser:     gmailUser,
		gmailAppPass:  gmailAppPass,
		mailFrom:      mailFrom,
		mailFromName:  mailFromName,
		subject:       subject,
		verifyBaseURL: verifyBaseURL,
	}
}

func (s *MailService) SendVerifyEmail(to string, token string) error {

	link := fmt.Sprintf("%s?token=%s",
		s.verifyBaseURL,
		url.QueryEscape(token),
	)

	htmlBody, err := s.renderVerifyTemplate(link)
	if err != nil {
		return err
	}

	fromHeader := fmt.Sprintf("%s <%s>", s.mailFromName, s.mailFrom)

	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", fromHeader),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", s.subject),
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
		"",
		htmlBody,
	}, "\r\n")

	log.Printf("[MAIL] smtp sending to=%s via=%s", to, "smtp.gmail.com:587")

	err = s.sendSMTPWithTimeout(to, []byte(msg))
	if err != nil {
		return err
	}

	log.Printf("[MAIL] sent to=%s", to)
	return nil
}

func (s *MailService) renderVerifyTemplate(link string) (string, error) {
	tmpl, err := template.ParseFiles("internal/templates/verify-email.html")
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"Link": link,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *MailService) sendSMTPWithTimeout(to string, msg []byte) error {
	addr := "smtp.gmail.com:587"

	// timeout ระดับ TCP
	conn, err := net.DialTimeout("tcp", addr, 8*time.Second)
	if err != nil {
		return err
	}
	// สำคัญ: กันค้างทั้ง connection
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))

	c, err := smtp.NewClient(conn, "smtp.gmail.com")
	if err != nil {
		return err
	}
	defer func() { _ = c.Quit() }()

	// STARTTLS
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: "smtp.gmail.com"}); err != nil {
			return err
		}
	}
	// Auth
	auth := smtp.PlainAuth("", s.gmailUser, s.gmailAppPass, "smtp.gmail.com")
	if err := c.Auth(auth); err != nil {
		return err
	}

	// From/To
	if err := c.Mail(s.mailFrom); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}
