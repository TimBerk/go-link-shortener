package logger

import (
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger = logrus.New()

type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (rl *responseLogger) WriteHeader(statusCode int) {
	rl.status = statusCode
	rl.w.WriteHeader(statusCode)
}

func (rl *responseLogger) Write(b []byte) (int, error) {
	size, err := rl.w.Write(b)
	rl.size += size
	return size, err
}

func (rl *responseLogger) Header() http.Header {
	return rl.w.Header()
}

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

func RequestLogger(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logrus.WithFields(logrus.Fields{
			"method": r.Method,
			"uri":    r.RequestURI,
		}).Info("Incoming request")

		rw := &responseLogger{w: w, status: http.StatusOK, size: 0}
		h(rw, r)

		logrus.WithFields(logrus.Fields{
			"status":   rw.status,
			"size":     rw.size,
			"duration": time.Since(start).String(),
		}).Info("Outgoing response")

	})
}
