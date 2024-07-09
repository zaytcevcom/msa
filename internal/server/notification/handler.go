package internalnotification

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
	r.HandleFunc("/notification/{userId}", h.GetByUserID).Methods(http.MethodGet)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(methodNotFoundHandler)

	return r
}

func (h *handler) Health(w http.ResponseWriter, r *http.Request) {
	response := h.app.Health(r.Context())
	writeResponseSuccess(w, response, h.logger)
}

func (h *handler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["userId"]

	if !ok {
		writeResponseError(w, fmt.Errorf("parameter 'userId' is missing from URL"), h.logger)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeResponseError(w, err, h.logger)
		return
	}

	notifications, err := h.app.GetByUserID(r.Context(), id)
	if err != nil {
		writeResponseError(w, err, h.logger)
		return
	}

	writeResponseSuccess(w, notifications, h.logger)
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
