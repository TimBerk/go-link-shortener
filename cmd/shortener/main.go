// Package main отвечает за конфигурирование, запуск и работу приложения
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/logger"
	"github.com/TimBerk/go-link-shortener/internal/app/router"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/TimBerk/go-link-shortener/internal/app/store/json"
	"github.com/TimBerk/go-link-shortener/internal/app/store/local"
	"github.com/TimBerk/go-link-shortener/internal/app/store/pg"
	"github.com/TimBerk/go-link-shortener/internal/app/workers"
	_ "github.com/TimBerk/go-link-shortener/swagger"
)

var (
	// buildVersion - версия сборки
	buildVersion string = "N/A"
	// buildDate - дата сборки
	buildDate string = "N/A"
	// buildCommit - коммит сборки
	buildCommit string = "N/A"
)

// printBuildInfo - выводит информацию о сборке
func printBuildInfo() {
	fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)
}

// @Title Shortener API
// @Description Сервис сокращения URL
// @Version 1.0

// @BasePath /api/v1
// @Host localhost:8080

func main() {
	printBuildInfo()

	cfg := config.InitConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errLogs := logger.Initialize(cfg.LogLevel)
	if errLogs != nil {
		logger.Log.Fatal("Error initializing logs: ", errLogs)
	}

	generator := store.NewIDGenerator()
	urlChan := make(chan store.URLPair, 1000)
	signalChan := make(chan os.Signal, 1)

	var dataStore store.Store
	var errStore error

	if cfg.DatabaseDSN != "" {
		dataStore, errStore = pg.NewPgStore(generator, cfg)
	} else if cfg.UseLocalStore {
		dataStore, errStore = local.NewURLStore(generator)
	} else {
		dataStore, errStore = json.NewJSONStore(cfg.FileStoragePath, generator)
	}
	if errStore != nil {
		logger.Log.Fatal("Read Store: ", errStore)
	}

	// Запускаем воркер
	var wg sync.WaitGroup
	wg.Add(1)
	go workers.Worker(ctx, dataStore, urlChan, &wg)

	// Запускаем воркер для сигналов
	wg.Add(1)
	go workers.SignalWorker(ctx, cancel, signalChan, &wg)

	router := router.RegisterRouters(dataStore, cfg, ctx, urlChan)

	serverErrChan := make(chan error, 1)
	go func() {
		logger.Log.WithField("address", cfg.ServerAddress).Info("Starting server")
		var errRun error
		if !cfg.EnableHTTPS {
			errRun = http.ListenAndServe(cfg.ServerAddress, router)
		} else {
			certFile := "cert.pem"
			if _, err := os.Stat(certFile); os.IsNotExist(err) {
				logger.Log.WithField("file", certFile).Fatal("Certificate file is not found", err)
			}

			keyFile := "key.pem"
			if _, err := os.Stat(keyFile); os.IsNotExist(err) {
				logger.Log.WithField("file", keyFile).Fatal("Key file is not found", err)
			}

			errRun = http.ListenAndServeTLS(cfg.ServerAddress, certFile, keyFile, router)
		}

		if errRun != nil && !errors.Is(errRun, http.ErrServerClosed) {
			serverErrChan <- errRun
		}
	}()

	// Основной цикл обработки
	select {
	case err := <-serverErrChan:
		logger.Log.Fatalf("Server error: %v", err)
	case <-ctx.Done():
		logger.Log.Info("Shutdown initiated by signal worker")

		// Закрываем канал и ждем завершения воркеров
		close(urlChan)
		wg.Wait()

		logger.Log.Info("Server shutdown completed")
	}
}
