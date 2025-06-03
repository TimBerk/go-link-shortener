// Package checker выступает в роли обработчика запросов с учетом доступа по IP
package checker

import (
	"net"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/TimBerk/go-link-shortener/internal/app/config"
	"github.com/TimBerk/go-link-shortener/internal/app/middlewares/logger"
)

// TrustedSubnetMiddleware - обработчик запросов для проверки доступности IP
func TrustedSubnetMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.TrustedSubnet == "" {
				logger.Log.WithField("TrustedSubnet", cfg.TrustedSubnet).Error("Empty subnet")
				http.Error(w, "Access forbidden", http.StatusForbidden)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				logger.Log.WithField("realIP", realIP).Error("Empty realIP")
				http.Error(w, "X-Real-IP header required", http.StatusForbidden)
				return
			}

			_, trustedNet, err := net.ParseCIDR(cfg.TrustedSubnet)
			if err != nil {
				logger.Log.WithField("trustedNet", trustedNet).Error("Invalid trusted subnet configuration")
				http.Error(w, "Invalid trusted subnet configuration", http.StatusInternalServerError)
				return
			}

			clientIP := net.ParseIP(realIP)
			if clientIP == nil {
				logger.Log.WithField("clientIP", clientIP).Error("Invalid IP address")
				http.Error(w, "Invalid IP address", http.StatusForbidden)
				return
			}

			if !trustedNet.Contains(clientIP) {
				logger.Log.WithFields(logrus.Fields{
					"clientIP":   clientIP,
					"trustedNet": trustedNet,
				}).Error("Invalid IP address for subnet")
				http.Error(w, "Access forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
