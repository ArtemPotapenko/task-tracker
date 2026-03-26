package app

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"task-tracker/shared-libs/pkg/logger"
)

type rateLimitMiddleware struct {
	burst   int
	limit   rate.Limit
	mu      sync.Mutex
	clients map[string]*clientLimiter
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newRateLimitMiddleware(rps, burst int) *rateLimitMiddleware {
	m := &rateLimitMiddleware{
		burst:   burst,
		limit:   rate.Limit(rps),
		clients: make(map[string]*clientLimiter),
	}
	go m.cleanupLoop()
	return m
}

func (m *rateLimitMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !m.allow(ip) {
			logger.Log.Infof("gateway rate limit: blocked ip=%s path=%s", ip, r.URL.Path)
			w.Header().Set("Retry-After", "1")
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *rateLimitMiddleware) allow(ip string) bool {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[ip]
	if !ok {
		client = &clientLimiter{
			limiter:  rate.NewLimiter(m.limit, m.burst),
			lastSeen: now,
		}
		m.clients[ip] = client
	}
	client.lastSeen = now
	return client.limiter.Allow()
}

func (m *rateLimitMiddleware) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for now := range ticker.C {
		m.mu.Lock()
		for ip, client := range m.clients {
			if now.Sub(client.lastSeen) > 5*time.Minute {
				delete(m.clients, ip)
			}
		}
		m.mu.Unlock()
	}
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	if r.RemoteAddr == "" {
		return "unknown"
	}
	return r.RemoteAddr
}
