package internalauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type handler struct {
	logger Logger
	app    Application
}

type LoginRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func NewHandler(logger Logger, app Application) http.Handler {
	h := &handler{
		logger: logger,
		app:    app,
	}

	r := mux.NewRouter()
	r.HandleFunc("/health", h.Health).Methods(http.MethodGet)
	r.HandleFunc("/auth", h.Auth).Methods(http.MethodGet)
	r.HandleFunc("/login", h.Login).Methods(http.MethodPost)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(methodNotFoundHandler)

	return r
}

func (s *handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *handler) Login(w http.ResponseWriter, r *http.Request) {
	var loginData LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
		return
	}

	token, err := s.app.Login(r.Context(), loginData.Username, loginData.Password)
	if err != nil {
		s.logger.Error(err.Error())
		http.Error(w, "401 Unauthorized", http.StatusUnauthorized)

		return
	}

	writeResponseSuccess(w, token, s.logger)
}

func (s *handler) Auth(w http.ResponseWriter, r *http.Request) {
	userID, err := s.app.Auth(r.Context(), r.Header)
	if err == nil {
		w.Header().Set("X-User-Id", strconv.Itoa(userID))
	}

	w.WriteHeader(http.StatusOK)
}

func methodNotAllowedHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
}

func methodNotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "404 Not Found", http.StatusNotFound)
}

func writeResponseSuccess(w http.ResponseWriter, data interface{}, logger Logger) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	buf, err := json.Marshal(data)
	if err != nil {
		logger.Error(fmt.Sprintf("response marshal error: %s", err))
	}
	_, err = w.Write(buf)
	if err != nil {
		logger.Error(fmt.Sprintf("response marshal error: %s", err))
	}
	w.WriteHeader(http.StatusOK)
}
