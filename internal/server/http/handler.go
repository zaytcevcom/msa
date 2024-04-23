package internalhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zaytcevcom/msa/internal/storage"
)

type handler struct {
	logger Logger
	app    Application
}

type UserRequest struct {
	Username  string `json:"username,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewHandler(logger Logger, app Application) http.Handler {
	h := &handler{
		logger: logger,
		app:    app,
	}

	r := mux.NewRouter()
	r.HandleFunc("/health", h.Health).Methods(http.MethodGet)
	r.HandleFunc("/health/", h.Health).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}", h.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/user", h.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}", h.UpdateUser).Methods(http.MethodPut)
	r.HandleFunc("/user/{id}", h.DeleteUser).Methods(http.MethodDelete)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(methodNotFoundHandler)

	return r
}

func (s *handler) Health(w http.ResponseWriter, r *http.Request) {
	response := s.app.Health(r.Context())
	writeResponseSuccess(w, response, s.logger)
}

func (s *handler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]

	if !ok {
		writeResponseError(w, fmt.Errorf("parameter 'id' is missing from URL"), s.logger)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	user, err := s.app.GetByID(r.Context(), id)
	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	writeResponseSuccess(w, user, s.logger)
}

func (s *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userData UserRequest
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	id, err := s.app.Create(
		r.Context(),
		userData.Username,
		userData.FirstName,
		userData.LastName,
		userData.Email,
		userData.Phone,
	)
	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	writeResponseSuccess(w, id, s.logger)
}

func (s *handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	var userData UserRequest
	err = json.NewDecoder(r.Body).Decode(&userData)

	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	user := storage.User{
		ID:        id,
		Username:  userData.Username,
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Email:     userData.Email,
		Phone:     userData.Phone,
	}

	err = s.app.Update(r.Context(), id, user)

	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	writeResponseSuccess(w, id, s.logger)
}

func (s *handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]

	if !ok {
		writeResponseError(w, fmt.Errorf("parameter 'id' is missing from URL"), s.logger)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	err = s.app.Delete(r.Context(), id)

	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	writeResponseSuccess(w, 1, s.logger)
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
}

func writeResponseError(w http.ResponseWriter, err error, logger Logger) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := &ErrorResponse{}
	response.Error.Message = err.Error()

	buf, err := json.Marshal(response)
	if err != nil {
		logger.Error(fmt.Sprintf("response marshal error: %s", err))
	}
	_, err = w.Write(buf)
	if err != nil {
		logger.Error(fmt.Sprintf("response marshal error: %s", err))
	}
}
