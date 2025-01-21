package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	LogLevel      string
}

func InitConfig() *Config {
	cfg := &Config{}

	envServerAddress := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envLogLevel := os.Getenv("LOGGING_LEVEL")

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Logging level")

	flag.Parse()

	if envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}
	if envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}
	if envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	return cfg
}

func NewConfig(serverAddress, baseURL string) *Config {
	return &Config{
		ServerAddress: serverAddress,
		BaseURL:       baseURL,
	}
}
