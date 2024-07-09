package ordermiddleware

import (
	"context"
	"net/http"
	"strconv"
)

type UserIDKey struct{}

type RequestIDKey struct{}

func HeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerUserID := r.Header.Get("X-User-Id")
		var userID int
		if headerUserID != "" {
			userID, _ = strconv.Atoi(headerUserID)
		}

		ctx := context.WithValue(r.Context(), UserIDKey{}, userID)

		requestID := r.Header.Get("X-Request-Id")
		ctx = context.WithValue(ctx, RequestIDKey{}, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
