// Package workers необходим для обработки фоновых задач
package workers

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/logger"
)

// SignalWorker в фоне получает сигналы от системы и логирует их
func SignalWorker(ctx context.Context, cancel context.CancelFunc, sigChan chan os.Signal, wg *sync.WaitGroup) {
	defer wg.Done()

	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case sig := <-sigChan:
		logger.Log.Infof("Received signal: %v. Initiating shutdown...", sig)
		cancel()
	case <-ctx.Done():
		return
	}
}
