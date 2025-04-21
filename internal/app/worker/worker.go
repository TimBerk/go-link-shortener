package worker

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/TimBerk/go-link-shortener/internal/app/store"
)

const (
	batchLimit = 100
)

func Worker(ctx context.Context, dataStore store.Store, urlChan <-chan store.URLPair, wg *sync.WaitGroup) {
	defer wg.Done()

	var batch []store.URLPair
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case pair, ok := <-urlChan:
			if !ok {
				if len(batch) > 0 {
					flushBatch(ctx, batch, dataStore)
				}
				return
			}
			batch = append(batch, pair)
			if len(batch) >= batchLimit {
				flushBatch(ctx, batch, dataStore)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				flushBatch(ctx, batch, dataStore)
				batch = nil
			}
		case <-ctx.Done():
			if len(batch) > 0 {
				newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				flushBatch(newCtx, batch, dataStore)
			}
			return
		}
	}
}

func flushBatch(ctx context.Context, batch []store.URLPair, dataStore store.Store) {
	if len(batch) == 0 {
		return
	}

	logrus.WithField("count", len(batch)).Info("Flush batch URLs")

	dataStore.DeleteURL(ctx, batch)
}
