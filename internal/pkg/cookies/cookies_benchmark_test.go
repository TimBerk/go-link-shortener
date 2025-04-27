package cookies

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkGenerateUserID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateUserID()
	}
}

func BenchmarkGetEncodedValue(b *testing.B) {
	userID := GenerateUserID()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetEncodedValue(userID)
	}
}

func BenchmarkSetUserCookie(b *testing.B) {
	userID := GenerateUserID()
	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetUserCookie(w, userID)
	}
}

func BenchmarkGetUserID(b *testing.B) {
	userID := GenerateUserID()
	encoded, _ := GetEncodedValue(userID)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: "user", Value: encoded})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetUserID(req)
	}
}

func BenchmarkFullCycle(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := GenerateUserID()
		SetUserCookie(w, userID)

		resp := w.Result()
		defer resp.Body.Close()

		cookies := resp.Cookies()

		if len(cookies) > 0 {
			req.AddCookie(cookies[0])
			_, _ = GetUserID(req)
		}

		w = httptest.NewRecorder()
		req = httptest.NewRequest("GET", "http://example.com", nil)
	}
}

func BenchmarkGetUserID_InvalidCookie(b *testing.B) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: "user", Value: "invalid"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetUserID(req)
	}
}

func BenchmarkGetUserID_NoCookie(b *testing.B) {
	req := httptest.NewRequest("GET", "http://example.com", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetUserID(req)
	}
}
