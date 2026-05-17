package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL             string
	JWTSecret               string
	AdminPassword           string
	Port                    int
	DashboardPort           int
	EnableImageGeneration   bool
	GitHubClientID          string
	GitHubClientSecret      string
	GeminiClientSecret      string
	IflowClientSecret       string
	AntigravityClientSecret string
	EncryptionKey           string
	AllowedOrigins          []string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ai_proxy?sslmode=disable"),
		JWTSecret:               getEnv("JWT_SECRET", "change-me-in-production"),
		AdminPassword:           getEnv("ADMIN_PASSWORD", "admin"),
		Port:                    getEnvInt("PORT", 20128),
		DashboardPort:           getEnvInt("DASHBOARD_PORT", 1433),
		EnableImageGeneration:   getEnvBool("ENABLE_IMAGE_GENERATION", false),
		GitHubClientID:          getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret:      getEnv("GITHUB_CLIENT_SECRET", ""),
		GeminiClientSecret:      getEnv("GEMINI_CLIENT_SECRET", ""),
		IflowClientSecret:       getEnv("IFLOW_CLIENT_SECRET", ""),
		AntigravityClientSecret: getEnv("ANTIGRAVITY_CLIENT_SECRET", ""),
		EncryptionKey:           getEnv("ENCRYPTION_KEY", ""),
		AllowedOrigins:          parseOrigins(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:1433")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

var cfg *Config

func InitConfig() *Config {
	cfg = Load()
	return cfg
}

func GetConfig() *Config {
	if cfg == nil {
		cfg = Load()
	}
	return cfg
}

func IsImageGenerationEnabled() bool {
	return GetConfig().EnableImageGeneration
}

func parseOrigins(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
