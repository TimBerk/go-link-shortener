package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	UseLocalStore   bool `envconfig:"USE_LOCAL_STORE" default:"false"`
	DatabaseDSN     string
}

func InitConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal("Failed to parse config: ", err)
	}

	envServerAddress := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envLogLevel := os.Getenv("LOGGING_LEVEL")
	envFileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	envUseLocalStore := os.Getenv("USE_LOCAL_STORE")
	envDatabaseDSN := os.Getenv("DATABASE_DSN")

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Logging level")
	flag.StringVar(&cfg.FileStoragePath, "p", "files/data.json", "Path for files")
	flag.BoolVar(&cfg.UseLocalStore, "local", false, "Use local store for url links")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database DSN for PostgreSQL")

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
	if envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}
	if envUseLocalStore != "" {
		boolVar, err := strconv.ParseBool(strings.ToLower(envUseLocalStore))
		if err != nil {
			boolVar = false
		}

		cfg.UseLocalStore = boolVar
	}
	if envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}

	return cfg
}

func NewConfig(serverAddress, baseURL string, useLocalStore bool) *Config {
	return &Config{
		ServerAddress: serverAddress,
		BaseURL:       baseURL,
		UseLocalStore: useLocalStore,
	}
}
