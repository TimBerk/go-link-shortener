// Package main отвечает за конфигурирование, запуск и работу приложения
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/crypto/acme/autocert"

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

	cfg := config.InitConfig()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

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

	serverErrChan := make(chan error, 1)
	go func() {
		logger.Log.WithField("address", cfg.ServerAddress).Info("Starting server")
		var errRun error
		if !cfg.EnableHTTPS {
			errRun = http.ListenAndServe(cfg.ServerAddress, router)
		} else {
			certManager := &autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist("test.com", "www.test.com"),
				Cache:      autocert.DirCache("certs"),
			}
			server := &http.Server{
				Addr:      ":443",
				Handler:   router,
				TLSConfig: certManager.TLSConfig(),
			}
			errRun = server.ListenAndServeTLS("", "")
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
