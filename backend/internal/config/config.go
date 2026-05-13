package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL             string
	JWTSecret               string
	AdminPassword           string
	Port                    int
	EnableImageGeneration   bool
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ai_proxy?sslmode=disable"),
		JWTSecret:               getEnv("JWT_SECRET", "change-me-in-production"),
		AdminPassword:           getEnv("ADMIN_PASSWORD", "admin"),
		Port:                    getEnvInt("PORT", 20128),
		EnableImageGeneration:   getEnvBool("ENABLE_IMAGE_GENERATION", false),
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
