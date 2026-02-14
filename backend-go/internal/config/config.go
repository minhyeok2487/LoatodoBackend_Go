package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	ServerPort string

	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// JWT
	JWTKey string

	// Lostark API
	LostarkAPIKey     string
	LostarkAPITestKey string
	APIKey3           string
	APIKey4           string

	// Discord Webhooks
	DiscordWebhookURL string
	DiscordNoticeURL  string
	DiscordMemberURL  string

	// Google OAuth2
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string

	// S3
	S3AccessKey  string
	S3SecretKey  string
	S3BucketName string
	S3Region     string

	// SMTP (Email)
	SMTPHost    string
	SMTPPort    string
	SMTPEmail   string
	SMTPPassword string

	// MyGame API Key
	MyGameAPIKey string

	// DB Pool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxIdleTime int // seconds
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "Loatodo"),
		DBUser:     getEnv("DB_USER", "admin"),
		DBPassword: getEnv("DB_PASSWORD", "admin1234"),

		JWTKey: getEnv("JWT_KEY", ""),

		LostarkAPIKey:     getEnv("LOSTARK_API_KEY", ""),
		LostarkAPITestKey: getEnv("LOSTARK_API_TEST_KEY", ""),
		APIKey3:           getEnv("API_KEY3", ""),
		APIKey4:           getEnv("API_KEY4", ""),

		DiscordWebhookURL: getEnv("DISCORD_WEBHOOK_URL", ""),
		DiscordNoticeURL:  getEnv("DISCORD_NOTICE_URL", ""),
		DiscordMemberURL:  getEnv("DISCORD_MEMBER_URL", ""),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI:  getEnv("GOOGLE_REDIRECT_URI", "https://api2.loatodo.com/login/oauth2/code/google"),

		S3AccessKey:  getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:  getEnv("S3_SECRET_KEY", ""),
		S3BucketName: getEnv("S3_BUCKET_NAME", "loatodo"),
		S3Region:     getEnv("S3_REGION", "ap-northeast-2"),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "465"),
		SMTPEmail:    getEnv("SMTP_EMAIL", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),

		MyGameAPIKey: getEnv("MYGAME_API_KEY", ""),

		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 20),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxIdleTime: getEnvInt("DB_CONN_MAX_IDLE_TIME", 60),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
