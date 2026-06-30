package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
)

type Config struct {
	Port      string
	Driver    string
	DSN       string
	JWTSecret string
}

func Load() *Config {
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			log.Fatalf("failed to generate JWT secret: %v", err)
		}
		jwtSecret = hex.EncodeToString(randomBytes)
		log.Println("WARNING: JWT_SECRET not set, using random secret. Set JWT_SECRET for persistent sessions across restarts.")
	}

	return &Config{
		Port:      getEnv("PORT", "8080"),
		Driver:    getEnv("DB_DRIVER", "sqlite"),
		DSN:       getEnv("DB_DSN", "./open-todo-server.db"),
		JWTSecret: jwtSecret,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
