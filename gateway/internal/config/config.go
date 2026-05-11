package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr        string
	GRPCBackendAddr string
	GRPCDialTimeout time.Duration
	ShutdownTimeout time.Duration

	LogLevel  string
	LogPretty bool

	CORSAllowedOrigins []string

	AuthEnabled    bool
	AuthJWTKey     string
	AuthAPIURL     string
	AuthAPITimeout time.Duration

	SwaggerEnabled bool
	SwaggerSpec    string
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		GRPCBackendAddr:    getEnv("GRPC_BACKEND_ADDR", "localhost:9090"),
		GRPCDialTimeout:    getDuration("GRPC_DIAL_TIMEOUT", 5*time.Second),
		ShutdownTimeout:    getDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		LogPretty:          getBool("LOG_PRETTY", false),
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "*")),
		AuthEnabled:        getBool("AUTH_ENABLED", false),
		AuthJWTKey:         getEnv("AUTH_JWT_KEY", ""),
		AuthAPIURL:         getEnv("AUTH_API_URL", ""),
		AuthAPITimeout:     getDuration("AUTH_API_TIMEOUT", 3*time.Second),
		SwaggerEnabled:     getBool("SWAGGER_ENABLED", true),
		SwaggerSpec:        getEnv("SWAGGER_SPEC", "docs/openapi.json"),
	}

	if cfg.AuthEnabled && cfg.AuthJWTKey == "" && cfg.AuthAPIURL == "" {
		return nil, errors.New("AUTH_ENABLED=true requires AUTH_JWT_KEY or AUTH_API_URL")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
