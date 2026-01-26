package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	ErrMissingConfigPath = errors.New("CONFIG_PATH is not set")
	ErrMissingJWTSecret  = errors.New("JWT_SECRET is not set")
	ErrInvalidJWTTTL     = errors.New("invalid jwt.ttl (use duration like 15m, 24h)")
)

type Config struct {
	HTTP struct {
		Addr string `yaml:"addr"`
	} `yaml:"http"`

	DB struct {
		Path string `yaml:"path"`
	} `yaml:"db"`

	JWT struct {
		TTLRaw string        `yaml:"ttl"`
		TTL    time.Duration `yaml:"-"`
		Secret string        `yaml:"-"`
	} `yaml:"jwt"`
}

// Load reads YAML from CONFIG_PATH and secrets from ENV.
// ENV overrides (optional): HTTP_ADDR, DB_PATH, JWT_TTL, JWT_SECRET.
func Load() (Config, error) {
	var cfg Config

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		return cfg, ErrMissingConfigPath
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}

	// Optional overrides from ENV
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		cfg.HTTP.Addr = v
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DB.Path = v
	}
	if v := os.Getenv("JWT_TTL"); v != "" {
		cfg.JWT.TTLRaw = v
	}

	ttl, err := time.ParseDuration(cfg.JWT.TTLRaw)
	if err != nil || ttl <= 0 {
		return cfg, ErrInvalidJWTTTL
	}
	cfg.JWT.TTL = ttl

	cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	if cfg.JWT.Secret == "" {
		return cfg, ErrMissingJWTSecret
	}
	// Sensible defaults if YAML left empty
	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8080"
	}
	if cfg.DB.Path == "" {
		cfg.DB.Path = "data/tasks.db"
	}
	// ttlStr := os.Getenv("JWT_TTL")
	// if ttlStr == "" {
	// 	ttlStr = "15m"
	// }

	// ttl, err := time.ParseDuration(ttlStr)
	// if err != nil {
	// 	log.Fatal("invalid JWT_TTL:", err)
	// }

	return cfg, nil
}
