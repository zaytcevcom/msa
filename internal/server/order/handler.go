package internalorder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	ordermiddleware "github.com/zaytcevcom/msa/internal/server/order/middleware"
)

type handler struct {
	logger Logger
	app    Application
}

type CreateRequest struct {
	ProductID int     `json:"productId,omitempty"`
	Sum       float64 `json:"sum,omitempty"`
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
	r.HandleFunc("/order/{id}", h.GetByID).Methods(http.MethodGet)
	r.HandleFunc("/order", h.Create).Methods(http.MethodPost)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(methodNotFoundHandler)

	return r
}

func (s *handler) Health(w http.ResponseWriter, r *http.Request) {
	response := s.app.Health(r.Context())
	writeResponseSuccess(w, response, s.logger)
}

func (s *handler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	order, err := s.app.GetByID(r.Context(), id)
	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	writeResponseSuccess(w, order, s.logger)
}

func (s *handler) Create(w http.ResponseWriter, r *http.Request) {
	var data CreateRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	value := r.Context().Value(ordermiddleware.UserIDKey{})
	if value == nil {
		forbidden(w)
		return
	}

	userID, ok := value.(int)
	if !ok {
		forbidden(w)
		return
	}

	// todo: Hardcoded
	id, err := s.app.Create(r.Context(), userID, data.ProductID, data.Sum, "mail@example.com")
	if err != nil {
		writeResponseError(w, err, s.logger)
		return
	}

	writeResponseSuccess(w, id, s.logger)
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

func forbidden(w http.ResponseWriter) {
	http.Error(w, "403 Forbidden", http.StatusForbidden)
}
