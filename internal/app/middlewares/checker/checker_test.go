package checker

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TimBerk/go-link-shortener/internal/app/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		trustedSubnet string
		realIP        string
		expectedCode  int
		expectedBody  string
	}{
		{
			name:          "no trusted subnet configured",
			trustedSubnet: "",
			realIP:        "192.168.1.1",
			expectedCode:  http.StatusForbidden,
			expectedBody:  "Access forbidden\n",
		},
		{
			name:          "missing X-Real-IP header",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "",
			expectedCode:  http.StatusForbidden,
			expectedBody:  "X-Real-IP header required\n",
		},
		{
			name:          "invalid trusted subnet format",
			trustedSubnet: "invalid_subnet",
			realIP:        "192.168.1.1",
			expectedCode:  http.StatusInternalServerError,
			expectedBody:  "Invalid trusted subnet configuration\n",
		},
		{
			name:          "invalid IP address",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "not_an_ip",
			expectedCode:  http.StatusForbidden,
			expectedBody:  "Invalid IP address\n",
		},
		{
			name:          "IP not in trusted subnet",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "10.0.0.1",
			expectedCode:  http.StatusForbidden,
			expectedBody:  "Access forbidden\n",
		},
		{
			name:          "IP in trusted subnet",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.100",
			expectedCode:  http.StatusOK,
			expectedBody:  "OK",
		},
		{
			name:          "IPv6 in trusted subnet",
			trustedSubnet: "2001:db8::/32",
			realIP:        "2001:db8::1",
			expectedCode:  http.StatusOK,
			expectedBody:  "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewConfig("localhost:8021", "http://base.url", true, tt.trustedSubnet)
			middleware := TrustedSubnetMiddleware(cfg)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}

func TestTrustedSubnetMiddlewareEdgeCases(t *testing.T) {
	t.Run("valid IP but not in subnet", func(t *testing.T) {
		_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
		testIP := "192.168.2.1"
		require.False(t, subnet.Contains(net.ParseIP(testIP)))
		cfg := config.NewConfig("localhost:8021", "http://base.url", true, "192.168.1.0/24")
		middleware := TrustedSubnetMiddleware(cfg)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Real-IP", testIP)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}
