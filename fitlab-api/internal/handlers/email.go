package handlers

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

// GetEmailConfig loads SMTP config from environment
func GetEmailConfig() EmailConfig {
	return EmailConfig{
		Host: os.Getenv("SMTP_HOST"),
		Port: os.Getenv("SMTP_PORT"),
		User: os.Getenv("SMTP_USER"),
		Pass: os.Getenv("SMTP_PASS"),
		From: os.Getenv("SMTP_FROM"),
	}
}

// SendRecoveryEmail sends password recovery email with temporary password
func SendRecoveryEmail(toEmail, tempPassword string) error {
	cfg := GetEmailConfig()

	if cfg.Host == "" || cfg.User == "" || cfg.Pass == "" {
		return fmt.Errorf("SMTP not configured")
	}

	// Sanitize email for From header
	fromName := "Fitlab"
	if cfg.From != "" {
		parts := strings.Split(cfg.From, "<")
		if len(parts) > 1 {
			fromName = strings.Trim(parts[0], " ")
		}
	}

	// Build email
	subject := "Recuperación de contraseña - Fitlab"
	body := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\n\r\nHola,\r\n\r\nTu contraseña temporal es: %s\r\n\r\nIngresá con esta contraseña y luego podés cambiarla desde tu perfil.\r\n\r\nSaludos,\r\nEquipo Fitlab\r\n", fromName, cfg.User, toEmail, subject, tempPassword)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	// Create TLS config that skips verification (for known hosts like Gmail)
	tlsConfig := &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: false,
	}

	// Connect and upgrade to TLS
	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("error connecting to SMTP: %v", err)
	}

	// Start TLS
	if err = conn.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("error starting TLS: %v", err)
	}

	// Authenticate
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	if err = conn.Auth(auth); err != nil {
		return fmt.Errorf("error authenticating: %v", err)
	}

	// Send email
	if err = conn.Mail(cfg.User); err != nil {
		return fmt.Errorf("error setting sender: %v", err)
	}

	if err = conn.Rcpt(toEmail); err != nil {
		return fmt.Errorf("error setting recipient: %v", err)
	}

	w, err := conn.Data()
	if err != nil {
		return fmt.Errorf("error creating data stream: %v", err)
	}

	_, err = w.Write([]byte(body))
	if err != nil {
		return fmt.Errorf("error writing email: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("error closing data stream: %v", err)
	}

	conn.Quit()

	return nil
}