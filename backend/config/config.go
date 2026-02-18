package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
)

type Config struct {
	Port         string
	AdminPIN     string
	JWTSecret    string
	DataDir      string
	FrontendDist string
}

func Load() *Config {
	cfg := &Config{
		Port:         getEnv("PORT", "8080"),
		AdminPIN:     getEnv("ADMIN_PIN", "1234"),
		JWTSecret:    getEnv("JWT_SECRET", ""),
		DataDir:      getEnv("DATA_DIR", "./data"),
		FrontendDist: getEnv("FRONTEND_DIST", "../frontend/dist"),
	}
	if cfg.JWTSecret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		cfg.JWTSecret = hex.EncodeToString(b)
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
