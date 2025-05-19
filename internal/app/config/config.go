// Package config работает с настройками для проекта.
// Осуществлена поддержка работы с переменными окружения и флагами.
package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/sirupsen/logrus"
)

// JSONFile структура для хранения json-конфигурации
type JSONFile struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// Config задает основные переменные окружения
type Config struct {
	ServerAddress   string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	UseLocalStore   bool `envconfig:"USE_LOCAL_STORE" default:"false"`
	DatabaseDSN     string
	EnableHTTPS     bool   `envconfig:"ENABLE_HTTPS" default:"false"`
	ConfigFile      string `envconfig:"CONFIG"`
}

// InitConfig Инициализирует и устанавливает значения для переменных окружения
func InitConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		logrus.Fatal("Failed to parse config: ", err)
	}

	envServerAddress := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envLogLevel := os.Getenv("LOGGING_LEVEL")
	envFileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	envUseLocalStore := os.Getenv("USE_LOCAL_STORE")
	envDatabaseDSN := os.Getenv("DATABASE_DSN")
	envEnableHTTPS := os.Getenv("ENABLE_HTTPS")
	envConfigFile := os.Getenv("CONFIG")

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Logging level")
	flag.StringVar(&cfg.FileStoragePath, "p", "files/data.json", "Path for files")
	flag.BoolVar(&cfg.UseLocalStore, "local", false, "Use local store for url links")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database DSN for PostgreSQL")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "Enable HTTPS server")
	flag.StringVar(&cfg.ConfigFile, "c", "", "path to JSON config for server")

	flag.Parse()

	cfgJSON := JSONFile{}
	if envConfigFile == "" {
		envConfigFile = os.Getenv("CONFIG")
	}
	if envConfigFile != "" {
		file, errReadFile := os.ReadFile(envConfigFile)
		if errReadFile != nil {
			logrus.Warning("Couldn't read config file", errReadFile)
		} else {
			if errConfigJSON := json.Unmarshal(file, &cfgJSON); errConfigJSON != nil {
				logrus.Warning("Couldn't parse config file", errConfigJSON)
			}
		}
	}

	if envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	} else if cfgJSON.ServerAddress != "" {
		cfg.ServerAddress = cfgJSON.ServerAddress
	}

	if envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	} else if cfgJSON.BaseURL != "" {
		cfg.BaseURL = cfgJSON.BaseURL
	}

	if envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	} else if cfgJSON.FileStoragePath != "" {
		cfg.FileStoragePath = cfgJSON.FileStoragePath
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
	} else if cfgJSON.DatabaseDSN != "" {
		cfg.DatabaseDSN = cfgJSON.DatabaseDSN
	}

	if envEnableHTTPS != "" {
		boolVar, err := strconv.ParseBool(strings.ToLower(envEnableHTTPS))
		if err != nil {
			boolVar = false
		}

		cfg.EnableHTTPS = boolVar
	} else if cfgJSON.EnableHTTPS {
		cfg.EnableHTTPS = cfgJSON.EnableHTTPS
	}

	return cfg
}

// NewConfig Инициализирует минимальные настройки
func NewConfig(serverAddress, baseURL string, useLocalStore bool) *Config {
	return &Config{
		ServerAddress: serverAddress,
		BaseURL:       baseURL,
		UseLocalStore: useLocalStore,
		EnableHTTPS:   false,
	}
}
