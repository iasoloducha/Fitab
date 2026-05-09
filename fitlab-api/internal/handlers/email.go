package handlers

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type EmailConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

func GetEmailConfig() EmailConfig {
	return EmailConfig{
		Host: os.Getenv("SMTP_HOST"),
		Port: os.Getenv("SMTP_PORT"),
		User: os.Getenv("SMTP_USER"),
		Pass: os.Getenv("SMTP_PASS"),
		From: os.Getenv("SMTP_FROM"),
	}
}

func (cfg EmailConfig) IsConfigured() bool {
	return cfg.Host != "" && cfg.User != "" && cfg.Pass != ""
}

func (cfg EmailConfig) FromName() string {
	if cfg.From == "" {
		return "Fitlab"
	}
	parts := strings.Split(cfg.From, "<")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[0])
	}
	return "Fitlab"
}

func (cfg EmailConfig) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	tlsConfig := &tls.Config{ServerName: cfg.Host}

	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("connecting to SMTP: %v", err)
	}
	defer conn.Close()

	if err = conn.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("starting TLS: %v", err)
	}

	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	if err = conn.Auth(auth); err != nil {
		return fmt.Errorf("authenticating: %v", err)
	}

	if err = conn.Mail(cfg.User); err != nil {
		return fmt.Errorf("setting sender: %v", err)
	}
	if err = conn.Rcpt(to); err != nil {
		return fmt.Errorf("setting recipient: %v", err)
	}

	w, err := conn.Data()
	if err != nil {
		return fmt.Errorf("creating data stream: %v", err)
	}

	fullBody := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		cfg.FromName(), cfg.User, to, subject, body)

	if _, err = w.Write([]byte(fullBody)); err != nil {
		w.Close()
		return fmt.Errorf("writing email: %v", err)
	}

	w.Close()
	return nil
}

func SendRecoveryEmail(toEmail, tempPassword string) error {
	cfg := GetEmailConfig()
	if !cfg.IsConfigured() {
		return fmt.Errorf("SMTP not configured")
	}

	body := fmt.Sprintf("Hola,\r\n\r\nTu contraseña temporal es: %s\r\n\r\nIngresá con esta contraseña y luego podés cambiarla desde tu perfil.\r\n\r\nSaludos,\r\nEquipo Fitlab\r\n", tempPassword)
	return cfg.Send(toEmail, "Recuperación de contraseña - Fitlab", body)
}

type ProgressStats struct {
	ExercisesLogged  int
	CompletionRate   float64
	PerExerciseData  string
}

func SendExpirationReminderEmail(studentEmail, studentName, professorEmail, routineTitle, endDate string) error {
	cfg := GetEmailConfig()
	if !cfg.IsConfigured() {
		return fmt.Errorf("SMTP not configured")
	}

	professorLine := ""
	if professorEmail != "" {
		professorLine = fmt.Sprintf("- Profesor: %s\r\n", professorEmail)
	}

	body := fmt.Sprintf("Hola %s,\r\n\r\nTu rutina \"%s\" está por vencer.\r\n\r\nFecha de fin: %s\r\n- Alumno: %s\r\n%s\r\nContactá a tu profesor para renovar tu rutina.\r\n\r\nSaludos,\r\nEquipo Fitlab\r\n",
		studentName, routineTitle, endDate, studentEmail, professorLine)

	if professorEmail != "" {
		profBody := fmt.Sprintf("Hola,\r\n\r\nLa rutina \"%s\" del alumno %s (%s) vence el %s.\r\n\r\nConsiderá contactar al alumno para renovar su rutina.\r\n\r\nSaludos,\r\nEquipo Fitlab\r\n",
			routineTitle, studentName, studentEmail, endDate)
		cfg.Send(professorEmail, "Recordatorio: rutina por vencer - Fitlab", profBody)
	}

	return cfg.Send(studentEmail, "Recordatorio: rutina por vencer - Fitlab", body)
}

func SendRoutineCompletedEmail(toEmail, studentName, professorName, routineTitle string, stats ProgressStats) error {
	cfg := GetEmailConfig()
	if !cfg.IsConfigured() {
		return fmt.Errorf("SMTP not configured")
	}

	body := fmt.Sprintf("Hola %s,\r\n\r\nTu rutina \"%s\" ha finalizado.\r\n\r\nResumen de tu progreso:\r\n- Ejercicios registrados: %d\r\n- Tasa de completitud: %.0f%%\r\n%s\r\n¡Gracias por entrenar con Fitlab!\r\n\r\nSaludos,\r\nEquipo Fitlab\r\n",
		studentName, routineTitle, stats.ExercisesLogged, stats.CompletionRate, stats.PerExerciseData)

	return cfg.Send(toEmail, "Tu rutina ha finalizado - Fitlab", body)
}
