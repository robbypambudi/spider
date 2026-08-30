package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/internal/service"
	"github.com/spider/spider/internal/spidererrors"
)

type contextKey string

const UserContextKey contextKey = "user"

func Auth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeError(w, spidererrors.Authentication("Missing bearer token"))
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")
			user, err := auth.UserFromToken(r.Context(), token)
			if err != nil {
				writeError(w, err)
				return
			}
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WorkerAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Spider-Worker-Token") != token {
				writeError(w, spidererrors.WorkerAuth("Invalid worker token"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserFromContext(ctx context.Context) (*store.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*store.User)
	return user, ok
}

func writeError(w http.ResponseWriter, err error) {
	if se, ok := err.(*spidererrors.SpiderError); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(se.StatusCode)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": se.Code, "message": se.Message})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "spider_error", "message": err.Error()})
}
