// Package logger обрабатывает логи для приложения
package logger

import (
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Log - сущность логгера
var Log *logrus.Logger = logrus.New()

// gzipWriter - параметры для работы с логгами
type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

// WriteHeader - запись данных в заголовок с помощью обработчика записи
func (rl *responseLogger) WriteHeader(statusCode int) {
	rl.status = statusCode
	rl.w.WriteHeader(statusCode)
}

// Write - запись данных с помощью обработчика
func (rl *responseLogger) Write(b []byte) (int, error) {
	size, err := rl.w.Write(b)
	rl.size += size
	return size, err
}

// Header - получение заголовка из обработчика записи
func (rl *responseLogger) Header() http.Header {
	return rl.w.Header()
}

// Initialize Инициализирует и устанавливает значения для логгов
func Initialize(level string) error {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetOutput(os.Stdout)
	Log.SetLevel(logLevel)
	return nil
}

// RequestLogger - обработчик, добавлеяющий логирование для входящих/выохдящих запросов
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logrus.WithFields(logrus.Fields{
			"method": r.Method,
			"uri":    r.RequestURI,
		}).Info("Incoming request")

		rw := &responseLogger{w: w, status: http.StatusOK, size: 0}
		next.ServeHTTP(rw, r)

		logrus.WithFields(logrus.Fields{
			"status":   rw.status,
			"size":     rw.size,
			"duration": time.Since(start).String(),
		}).Info("Outgoing response")

	})
}
