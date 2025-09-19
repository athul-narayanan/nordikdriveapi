package config

import (
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	GeminiKey  string
	SMTPKey    string
	GmailUser  string
	GmailPass  string
}

func LoadConfig() Config {
	return Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		GeminiKey:  os.Getenv("GEMINI_KEY"),
		GmailUser:  os.Getenv("GMAIL_USER"),
		GmailPass:  os.Getenv("GMAIL_APP_PASSWORD"),
	}
}
