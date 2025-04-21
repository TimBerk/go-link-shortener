package utils

import (
	"github.com/sirupsen/logrus"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkCheckParamInHeaderParam(b *testing.B) {
	logrus.SetLevel(logrus.PanicLevel)
	testCases := []struct {
		name        string
		headerName  string
		headerValue string
		needParam   string
		expected    bool
	}{
		{
			name:        "ExactMatch",
			headerName:  "Content-Type",
			headerValue: "application/json",
			needParam:   "application/json",
			expected:    true,
		},
		{
			name:        "MatchWithExtraParams",
			headerName:  "Content-Type",
			headerValue: "application/json; charset=utf-8",
			needParam:   "application/json",
			expected:    true,
		},
		{
			name:        "NoMatch",
			headerName:  "Content-Type",
			headerValue: "text/html",
			needParam:   "application/json",
			expected:    false,
		},
		{
			name:        "EmptyHeader",
			headerName:  "Content-Type",
			headerValue: "",
			needParam:   "application/json",
			expected:    false,
		},
		{
			name:        "MalformedHeader",
			headerName:  "Content-Type",
			headerValue: ";",
			needParam:   "application/json",
			expected:    false,
		},
		{
			name:        "CaseSensitiveCheck",
			headerName:  "Content-Type",
			headerValue: "APPLICATION/JSON",
			needParam:   "application/json",
			expected:    false,
		},
		{
			name:        "WhitespaceCheck",
			headerName:  "Content-Type",
			headerValue: "  application/json  ",
			needParam:   "application/json",
			expected:    true,
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			if tc.headerValue != "" {
				req.Header.Set(tc.headerName, tc.headerValue)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = CheckParamInHeaderParam(req, tc.headerName, tc.needParam)
			}
		})
	}
}

func BenchmarkCheckParamInHeaderParam_HeaderSize(b *testing.B) {
	logrus.SetLevel(logrus.PanicLevel)
	headerSizes := []int{1, 10, 100, 1000}
	headerName := "Content-Type"
	needParam := "application/json"

	for _, size := range headerSizes {
		b.Run("Size is symbols "+strconv.Itoa(size), func(b *testing.B) {
			headerValue := "application/json; " + strings.Repeat("a=", size/2) + strings.Repeat("b", size%2)

			req := httptest.NewRequest("GET", "http://example.com", nil)
			req.Header.Set(headerName, headerValue)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = CheckParamInHeaderParam(req, headerName, needParam)
			}
		})
	}
}
