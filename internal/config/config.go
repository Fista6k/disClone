package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr           string
	JWTSecret      []byte
	OriginPatterns []string
}

func Load() Config {
	return Config{
		Addr:           getenv("ADDR", ":8080"),
		JWTSecret:      []byte(mustenv("JWT_SECRET")),
		OriginPatterns: listEnv("WS_ORIGIN_PATTERNS"),
	}
}

func getenv(key string, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return def
}

func mustenv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("missing env: " + key)
	}
	return value
}

func listEnv(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}

	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}

	return values
}
