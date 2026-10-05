package config

import "os"

type Config struct {
	Addr      string
	JWTSecret []byte
}

func Load() Config {
	return Config{
		Addr:      getenv("ADDR", ":8080"),
		JWTSecret: []byte(mustenv("JWT_SECRET")),
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
