// internal/middleware/middleware.go
package middleware

import (
	"context"
	"net/http"
	"option-manager/internal/service"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	UserIDKey  contextKey = "user_id"
	SessionKey contextKey = "session"
)

// Middleware represents a middleware function
type Middleware func(http.Handler) http.Handler

// Chain applies multiple middlewares to a handler
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

// RequireAuth ensures the user is authenticated
func RequireAuth(services *service.Services) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			session, err := services.Auth.GetSession(r.Context(), cookie.Value)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Add session and user info to context
			ctx := context.WithValue(r.Context(), SessionKey, session)
			ctx = context.WithValue(ctx, UserIDKey, session.UserID)

			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}

// GetSession retrieves the session from the context
func GetSession(ctx context.Context) (interface{}, bool) {
	session, ok := ctx.Value(SessionKey).(interface{})
	return session, ok
}

// You can add more middleware functions here:

// Recoverer handles panics
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
