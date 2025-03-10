package main

import (
	"context"
	"net/http"
	"sync"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/logger"
	"github.com/TimBerk/go-link-shortener/internal/app/router"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/TimBerk/go-link-shortener/internal/app/store/json"
	"github.com/TimBerk/go-link-shortener/internal/app/store/local"
	"github.com/TimBerk/go-link-shortener/internal/app/store/pg"
	"github.com/TimBerk/go-link-shortener/internal/app/worker"
)

func main() {
	ctx := context.Background()
	cfg := config.InitConfig()
	logger.Initialize(cfg.LogLevel)
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
