package mailer

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTPMailer struct {
	host    string
	port    string
	user    string
	pass    string
	from    string
	useTLS  bool
	timeout time.Duration
}

func NewSMTPMailer(host, port, user, pass, from string, useTLS bool, timeout time.Duration) (*SMTPMailer, error) {
	if strings.TrimSpace(host) == "" || strings.TrimSpace(port) == "" || strings.TrimSpace(from) == "" {
		return nil, errors.New("smtp host, port and from are required")
	}
	return &SMTPMailer{host: host, port: port, user: user, pass: pass, from: from, useTLS: useTLS, timeout: timeout}, nil
}

func (m *SMTPMailer) Send(to, subject, body string) error {
	addr := net.JoinHostPort(m.host, m.port)
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.from, to, subject, body))

	if !m.useTLS {
		return smtp.SendMail(addr, auth, m.from, []string{to}, msg)
	}

	dialer := net.Dialer{Timeout: m.timeout}
	conn, err := tls.DialWithDialer(&dialer, "tcp", addr, &tls.Config{ServerName: m.host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(m.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}
