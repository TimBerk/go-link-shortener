package worker

import (
	"context"
	"github.com/TimBerk/go-link-shortener/internal/app/store"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	batchLimit = 100
)

func Worker(ctx context.Context, dataStore store.Store, urlChan <-chan store.URLPair, wg *sync.WaitGroup) {
	defer wg.Done()

	var batch []store.URLPair
	timer := time.NewTimer(5 * time.Second)

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
				timer.Reset(5 * time.Second)
			}
		case <-timer.C:
			if len(batch) > 0 {
				flushBatch(ctx, batch, dataStore)
				batch = nil
			}
			timer.Reset(5 * time.Second)
		case <-ctx.Done():
			if len(batch) > 0 {
				flushBatch(ctx, batch, dataStore)
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
