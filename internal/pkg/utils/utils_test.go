package utils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

type mockCloser struct {
	closeError error
}

func (m *mockCloser) Close() error {
	return m.closeError
}

type failingResponseWriter struct {
	headers http.Header
	status  int
}

func (f *failingResponseWriter) Header() http.Header {
	return f.headers
}

func (f *failingResponseWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func (f *failingResponseWriter) WriteHeader(statusCode int) {
	f.status = statusCode
}

func TestCheckParamInHeaderParam(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		headerParam string
		needParam   string
		want        bool
	}{
		{
			name:        "Empty header",
			headers:     map[string]string{},
			headerParam: "Content-type",
			needParam:   "text/plain",
			want:        false,
		},
		{
			name:        "Header without necessary header param",
			headers:     map[string]string{"x-frame-options": "SAMEORIGIN"},
			headerParam: "Content-type",
			needParam:   "text/plain",
			want:        false,
		},
		{
			name:        "Header with necessary param",
			headers:     map[string]string{"Content-type": "text/plain;charset=utf-8"},
			headerParam: "Content-type",
			needParam:   "text/plain",
			want:        true,
		},
		{
			name:        "Header without necessary param",
			headers:     map[string]string{"Content-type": "charset=utf-8"},
			headerParam: "Content-type",
			needParam:   "text/plain",
			want:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			assert.Equal(
				t,
				CheckParamInHeaderParam(req, test.headerParam, test.needParam),
				test.want,
				"CheckParamInHeaderParam(%s, %s) mast return %t",
				test.headerParam,
				test.needParam,
				test.want,
			)

		})
	}

}

func TestWriteJSONError(t *testing.T) {
	hook := test.NewGlobal()

	tests := []struct {
		name         string
		message      string
		statusCode   int
		wantLogCount int
	}{
		{
			name:         "successful error response",
			message:      "Not found",
			statusCode:   404,
			wantLogCount: 1,
		},
		{
			name:         "empty message",
			message:      "",
			statusCode:   500,
			wantLogCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()

			w := httptest.NewRecorder()
			WriteJSONError(w, tt.message, tt.statusCode)

			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, `{"error":"`+tt.message+`"}`, w.Body.String())
			assert.Len(t, hook.Entries, tt.wantLogCount)
			if tt.wantLogCount > 0 {
				assert.Equal(t, logrus.ErrorLevel, hook.Entries[0].Level)
				assert.Contains(t, hook.Entries[0].Message, "JSON Error")
				assert.Equal(t, tt.message, hook.Entries[0].Data["message"])
				assert.Equal(t, tt.statusCode, hook.Entries[0].Data["statusCode"])
			}
		})
	}

	t.Run("failed json encoding", func(t *testing.T) {
		hook.Reset()
		w := &failingResponseWriter{
			headers: make(http.Header),
		}

		WriteJSONError(w, "write failure", 500)

		require.Len(t, hook.Entries, 2)
		assert.Equal(t, "JSON Error", hook.Entries[0].Message)
		assert.Equal(t, "Error creation response", hook.Entries[1].Message)
	})
}

func TestCloseWithLog(t *testing.T) {
	hook := test.NewGlobal()

	tests := []struct {
		name       string
		closeError error
		message    string
		wantLogged bool
		wantPanic  bool
	}{
		{
			name:       "successful close",
			closeError: nil,
			message:    "test closer",
			wantLogged: false,
		},
		{
			name:       "failed close",
			closeError: io.ErrUnexpectedEOF,
			message:    "test closer",
			wantLogged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()

			closer := &mockCloser{closeError: tt.closeError}
			CloseWithLog(closer, tt.message)

			if tt.wantLogged {
				require.Len(t, hook.Entries, 1)
				assert.Equal(t, logrus.ErrorLevel, hook.Entries[0].Level)
				assert.Equal(t, "Error closing", hook.Entries[0].Message)
				assert.Equal(t, tt.message, hook.Entries[0].Data["message"])
				assert.Equal(t, tt.closeError, hook.Entries[0].Data["error"])
			} else {
				assert.Empty(t, hook.Entries)
			}
		})
	}
}
