// Package main отвечает за конфигурирование, запуск и работу приложения
package main

import (
	"context"
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
	"github.com/TimBerk/go-link-shortener/internal/app/worker"
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

	ctx := context.Background()
	cfg := config.InitConfig()

	errLogs := logger.Initialize(cfg.LogLevel)
	if errLogs != nil {
		logger.Log.Fatal("Error initializing logs: ", errLogs)
	}

	generator := store.NewIDGenerator()
	urlChan := make(chan store.URLPair, 1000)

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
	go worker.Worker(ctx, dataStore, urlChan, &wg)

	router := router.RegisterRouters(dataStore, cfg, ctx, urlChan)
	logger.Log.WithField("address", cfg.ServerAddress).Info("Starting server")
	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		logger.Log.Fatal("ListenAndServe: ", err)
	}

	// Закрываем канал и ждем завершения воркера
	close(urlChan)
	wg.Wait()
}
