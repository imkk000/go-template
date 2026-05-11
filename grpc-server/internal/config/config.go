package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	GRPCAddr          string
	ReflectionEnabled bool
	ShutdownTimeout   time.Duration

	LogLevel  string
	LogPretty bool
}

func Load() *Config {
	return &Config{
		GRPCAddr:          getEnv("GRPC_ADDR", ":9090"),
		ReflectionEnabled: getBool("GRPC_REFLECTION", true),
		ShutdownTimeout:   getDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogPretty:         getBool("LOG_PRETTY", false),
	}
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
