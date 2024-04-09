package internalhttp

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type handler struct {
	logger Logger
	app    Application
}

func NewHandler(logger Logger, app Application) http.Handler {
	h := &handler{
		logger: logger,
		app:    app,
	}

	r := mux.NewRouter()
	r.HandleFunc("/health", h.Health).Methods(http.MethodGet)
	r.HandleFunc("/health/", h.Health).Methods(http.MethodGet)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(methodNotFoundHandler)

	return r
}

func (s *handler) Health(w http.ResponseWriter, r *http.Request) {
	message := s.app.Health(r.Context())

	if _, err := fmt.Fprint(w, message); err != nil {
		return
	}
}

func methodNotAllowedHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
}

func methodNotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "404 Not Found", http.StatusNotFound)
}
