package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func InitConfig() *Config {
	cfg := &Config{}

	envServerAddress := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")

	flag.Parse()

	if envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}
	if envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	return cfg
}

func NewConfig(serverAddress, baseURL string) *Config {
	return &Config{
		ServerAddress: serverAddress,
		BaseURL:       baseURL,
	}
}
