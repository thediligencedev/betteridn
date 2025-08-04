package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPassword             string
	DBName                 string
	DBSSLMode              string
	ServerPort             string
	ServerEnv              string
	SessionSecret          string
	SessionExpiry          time.Duration
	GoogleClientID         string
	GoogleClientSecret     string
	GoogleOAuthRedirectURL string
	FrontendURL            string
	SMTPHost               string
	SMTPPort               string
	SMTPFrom               string
	SMTPUser               string
	SMTPPass               string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	sessionExpiry, err := time.ParseDuration(os.Getenv("SESSION_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_EXPIRY value: %w", err)
	}

	return &Config{
		DBHost:                 os.Getenv("DB_HOST"),
		DBPort:                 os.Getenv("DB_PORT"),
		DBUser:                 os.Getenv("DB_USER"),
		DBPassword:             os.Getenv("DB_PASSWORD"),
		DBName:                 os.Getenv("DB_NAME"),
		DBSSLMode:              os.Getenv("DB_SSLMODE"),
		ServerPort:             os.Getenv("SERVER_PORT"),
		ServerEnv:              os.Getenv("SERVER_ENV"),
		SessionSecret:          os.Getenv("SESSION_SECRET"),
		SessionExpiry:          sessionExpiry,
		GoogleClientID:         os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:     os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleOAuthRedirectURL: os.Getenv("GOOGLE_REDIRECT_URL"),
		FrontendURL:            os.Getenv("FRONTEND_URL"),
		SMTPHost:               os.Getenv("SMTP_HOST"),
		SMTPPort:               os.Getenv("SMTP_PORT"),
		SMTPFrom:               os.Getenv("SMTP_FROM"),
		SMTPUser:               os.Getenv("SMTP_USER"),
		SMTPPass:               os.Getenv("SMTP_PASS"),
	}, nil
}

func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}
