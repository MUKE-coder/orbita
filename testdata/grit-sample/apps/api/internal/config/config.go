package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() (*Config, error) {
	c := &Config{DatabaseURL: os.Getenv("DATABASE_URL"), Port: os.Getenv("PORT")}
	if c.Port == "" {
		c.Port = "8080"
	}
	return c, nil
}
