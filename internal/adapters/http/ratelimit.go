package http

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for requests
type RateLimiter struct {
	ips      map[string]*rate.Limiter
	mu       sync.Mutex
	r        rate.Limit
	b        int
	lastSeen map[string]time.Time // To cleanup old entries
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		ips:      make(map[string]*rate.Limiter),
		r:        rate.Limit(rps),
		b:        burst,
		lastSeen: make(map[string]time.Time),
	}

	// Simple background cleanup every minute
	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.ips[ip] = limiter
	}

	rl.lastSeen[ip] = time.Now()
	return limiter
}

func (rl *RateLimiter) cleanupLoop() {
	for {
		time.Sleep(1 * time.Minute)
		rl.mu.Lock()
		for ip, t := range rl.lastSeen {
			if time.Since(t) > 3*time.Minute {
				delete(rl.ips, ip)
				delete(rl.lastSeen, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// If we can't parse IP, we might want to log it but let it pass or block.
			// For now, let's try X-Forwarded-For if behind proxy, but that's spoofable unless trusted.
			// We'll fall back to "unknown" which shares a bucket (bad) or just skip (risky).
			// Let's rely on RemoteAddr primarily for standard setups, or just use the whole string if split fails.
			ip = r.RemoteAddr
		}

		// Cloudflare support
		if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
			ip = cfIP
		} else if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// XFF can be comma separated, take first
			// But note: XFF is spoofable if not behind a trusted proxy that strips it.
			// Assuming standard setup where edge is trusted.
		}

		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
