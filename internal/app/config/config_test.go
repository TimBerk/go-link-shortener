package config

import (
	"flag"
	"os"
	"testing"

	"github.com/TimBerk/go-link-shortener/internal/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	tests := []struct {
		name           string
		envServerAddr  string
		envBaseURL     string
		flagServerAddr string
		flagBaseURL    string
		expectedAddr   string
		expectedBase   string
	}{
		{
			name:           "Default values",
			envServerAddr:  "",
			envBaseURL:     "",
			flagServerAddr: "",
			flagBaseURL:    "",
			expectedAddr:   "localhost:8080",
			expectedBase:   "http://localhost:8080",
		},
		{
			name:           "Environment variables",
			envServerAddr:  "localhost:8081",
			envBaseURL:     "http://localhost:8081",
			flagServerAddr: "",
			flagBaseURL:    "",
			expectedAddr:   "localhost:8081",
			expectedBase:   "http://localhost:8081",
		},
		{
			name:           "Flags",
			envServerAddr:  "",
			envBaseURL:     "",
			flagServerAddr: "localhost:8082",
			flagBaseURL:    "http://localhost:8082",
			expectedAddr:   "localhost:8082",
			expectedBase:   "http://localhost:8082",
		},
		{
			name:           "Environment variables override flags",
			envServerAddr:  "localhost:8083",
			envBaseURL:     "http://localhost:8083",
			flagServerAddr: "localhost:8084",
			flagBaseURL:    "http://localhost:8084",
			expectedAddr:   "localhost:8083",
			expectedBase:   "http://localhost:8083",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
			os.Args = []string{"cmd"}

			utils.SetENVWithLog("SERVER_ADDRESS", test.envServerAddr)
			utils.SetENVWithLog("BASE_URL", test.envBaseURL)

			if test.flagServerAddr != "" {
				os.Args = append(os.Args, "-a", test.flagServerAddr)
			}
			if test.flagBaseURL != "" {
				os.Args = append(os.Args, "-b", test.flagBaseURL)
			}

			cfg := InitConfig()

			assert.Equal(t, test.expectedAddr, cfg.ServerAddress, "Неверный адрес сервера")
			assert.Equal(t, test.expectedBase, cfg.BaseURL, "Неверный базовый URL")

			utils.UnsetENVWithLog("SERVER_ADDRESS")
			utils.UnsetENVWithLog("BASE_URL")
		})
	}
}
