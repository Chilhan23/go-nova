package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName          string
	AppPort          string
	AppEnv           string
	BaseURL          string
	DatabaseURL      string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	TelegramBotToken string
	TelegramChatID   string
	AIProvider       string
	AIEndpoint       string
	AIAPIKey         string
	AIModel          string
	MaxUploadImageMB int
	MaxUploadVideoMB int
	UploadDir        string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	maxImg, _ := strconv.Atoi(getEnv("MAX_UPLOAD_IMAGE_MB", "5"))
	maxVid, _ := strconv.Atoi(getEnv("MAX_UPLOAD_VIDEO_MB", "15"))

	return &Config{
		AppName:          getEnv("APP_NAME", "tech-nova"),
		AppPort:          getEnv("APP_PORT", "8080"),
		AppEnv:           getEnv("APP_ENV", "development"),
		BaseURL:          getEnv("BASE_URL", "http://localhost:8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5432/technova_db?sslmode=disable"),
		DBHost:           getEnv("DB_HOST", "127.0.0.1"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBName:           getEnv("DB_NAME", "technova_db"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:   getEnv("TELEGRAM_CHAT_ID", ""),
		AIProvider:       getEnv("AI_PROVIDER", "openrouter"),
		AIEndpoint:       getEnv("AI_ENDPOINT", "https://openrouter.ai/api/v1/chat/completions"),
		AIAPIKey:         getEnv("AI_API_KEY", ""),
		AIModel:          getEnv("AI_MODEL", "openai/gpt-4o-mini"),
		MaxUploadImageMB: maxImg,
		MaxUploadVideoMB: maxVid,
		UploadDir:        getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
