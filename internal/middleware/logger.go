// internal/middleware/logger.go
package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// sensitiveHeaders contains headers that should be masked in logs
var sensitiveHeaders = map[string]bool{
	"Authorization":   true,
	"Cookie":          true,
	"Set-Cookie":      true,
	"X-CSRF-Token":    true,
	"X-Session-Token": true,
}

// sensitiveParams contains URL parameters that should be masked in logs
var sensitiveParams = map[string]bool{
	"password":   true,
	"token":      true,
	"key":        true,
	"secret":     true,
	"credential": true,
}

// Logger logs request details
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture the status code
		rw := newResponseWriter(w)

		// Defer logging until after the request is processed
		defer func() {
			// Recover from panics and log them
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v\n%s", err, debug.Stack())
				rw.statusCode = http.StatusInternalServerError
			}

			duration := time.Since(start)

			// Clean the URL query parameters
			cleanedURL := cleanURL(r.URL.String())

			// Log the request details
			log.Printf(
				"[%s] %s %s - Status: %d - Duration: %s - IP: %s - User-Agent: %s",
				r.Method,
				r.Host,
				cleanedURL,
				rw.statusCode,
				duration.Round(time.Millisecond),
				getClientIP(r),
				r.UserAgent(),
			)

			// Log headers (excluding sensitive ones)
			if shouldLogHeaders() {
				logHeaders(r)
			}
		}()

		// Process the request
		next.ServeHTTP(rw, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// cleanURL masks sensitive information in URL parameters
func cleanURL(url string) string {
	// Split URL into path and query
	parts := strings.Split(url, "?")
	if len(parts) == 1 {
		return url
	}

	path := parts[0]
	query := parts[1]

	// Process each query parameter
	params := strings.Split(query, "&")
	for i, param := range params {
		kv := strings.Split(param, "=")
		if len(kv) != 2 {
			continue
		}

		key := strings.ToLower(kv[0])
		if sensitiveParams[key] {
			params[i] = fmt.Sprintf("%s=[REDACTED]", kv[0])
		}
	}

	return fmt.Sprintf("%s?%s", path, strings.Join(params, "&"))
}

// getClientIP extracts the client's IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// logHeaders logs non-sensitive headers
func logHeaders(r *http.Request) {
	for name, values := range r.Header {
		if sensitiveHeaders[name] {
			log.Printf("Header %s: [REDACTED]", name)
			continue
		}
		log.Printf("Header %s: %s", name, strings.Join(values, ", "))
	}
}

// shouldLogHeaders returns true if detailed header logging should be enabled
// This could be controlled by environment variables or other configuration
func shouldLogHeaders() bool {
	// For development environments, you might want to return true
	// For production, you might want to make this configurable
	return false
}
