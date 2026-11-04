package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

type Config struct {
	JWTSecret    string
	JWTExpiresIn time.Duration
	Redis        RedisConfig
}

// Load читает настройки из переменных окружения (.env уже загружен в main).
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET не задан в .env")
	}

	expires := os.Getenv("JWT_EXPIRES_IN")
	if expires == "" {
		expires = "1h"
	}
	d, err := time.ParseDuration(expires)
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRES_IN должен быть длительностью (например 1h): %w", err)
	}
	cfg.JWTExpiresIn = d

	cfg.Redis.Host = os.Getenv("REDIS_HOST")
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}
	portStr := os.Getenv("REDIS_PORT")
	if portStr == "" {
		portStr = "6379"
	}
	cfg.Redis.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("REDIS_PORT должен быть числом: %w", err)
	}
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")

	return cfg, nil
}
