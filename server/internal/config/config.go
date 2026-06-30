package config

import "os"

type Config struct {
	Port      string
	Driver    string
	DSN       string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		Driver:    getEnv("DB_DRIVER", "sqlite"),
		DSN:       getEnv("DB_DSN", "./open-todo-server.db"),
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
