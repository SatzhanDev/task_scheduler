package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	JWTSecret string
	JWTTTL    time.Duration
}

func Load() Config {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	ttlStr := os.Getenv("JWT_TTL")
	if ttlStr == "" {
		ttlStr = "15m"
	}

	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		log.Fatal("invalid JWT_TTL:", err)
	}

	return Config{
		JWTSecret: secret,
		JWTTTL:    ttl,
	}
}
