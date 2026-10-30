package storage

import (
	"os"
)

type Config struct {
	DatabaseURL string
	Addr        string
}

func LoadConfig() Config {
	addr := os.Getenv("STORAGE_ADDR")
	if addr == "" {
		addr = ":8081"
	}
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Addr:        addr,
	}
}
