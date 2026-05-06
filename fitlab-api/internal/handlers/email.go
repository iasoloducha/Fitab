package handlers

import (
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
	body := fmt.Sprintf(`From: %s <%s>
To: %s
Subject: %s

Hola,

Tu contraseña temporal es: %s

Ingresá con esta contraseña y luego podés cambiarla desde tu perfil.

Si no solicitaste este cambio, ignorá este email.

Saludos,
Equipo Fitlab
`, fromName, cfg.User, toEmail, subject, tempPassword)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	
	// Use TLS for connection
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)

	// Send email
	err := smtp.SendMail(addr, auth, cfg.User, []string{toEmail}, []byte(body))
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	return nil
}