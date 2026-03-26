package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitMiddlewareBlocksBurst(t *testing.T) {
	middleware := newRateLimitMiddleware(1, 1)
	handler := middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "127.0.0.1:12345"
	resp1 := httptest.NewRecorder()
	handler.ServeHTTP(resp1, req1)
	if resp1.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want %d", resp1.Code, http.StatusOK)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "127.0.0.1:12345"
	resp2 := httptest.NewRecorder()
	handler.ServeHTTP(resp2, req2)
	if resp2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want %d", resp2.Code, http.StatusTooManyRequests)
	}
}

func TestClientIPUsesForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.1")

	if got := clientIP(req); got != "203.0.113.10" {
		t.Fatalf("clientIP() = %q, want %q", got, "203.0.113.10")
	}
}
