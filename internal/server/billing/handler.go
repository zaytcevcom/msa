package internalbilling

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	billingmiddleware "github.com/zaytcevcom/msa/internal/server/billing/middleware"
)

type handler struct {
	logger Logger
	app    Application
}

type AccountRequest struct {
	UserID int `json:"userId,omitempty"`
}

type DepositRequest struct {
	Amount float64 `json:"amount,omitempty"`
}

type WithdrawRequest struct {
	OrderID int     `json:"orderId,omitempty"`
	Amount  float64 `json:"amount,omitempty"`
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
	r.HandleFunc("/account", h.CreateAccount).Methods(http.MethodPost)
	r.HandleFunc("/account", h.GetAccount).Methods(http.MethodGet)
	r.HandleFunc("/account/{id}/deposit", h.Deposit).Methods(http.MethodPost)
	r.HandleFunc("/account/{id}/withdraw", h.Withdraw).Methods(http.MethodPost)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(methodNotFoundHandler)

	return r
}

func (s *handler) Health(w http.ResponseWriter, r *http.Request) {
	response := s.app.Health(r.Context())
	writeResponseSuccess(w, response, s.logger)
}

func (s *handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var data AccountRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeResponseError(w, err, s.logger, http.StatusBadGateway)
		return
	}

	id, err := s.app.CreateAccount(r.Context(), data.UserID)
	if err != nil {
		writeResponseError(w, err, s.logger, http.StatusBadGateway)
		return
	}

	writeResponseSuccess(w, id, s.logger)
}

func (s *handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(billingmiddleware.UserIDKey{})
	if value == nil {
		forbidden(w)
		return
	}

	id, ok := value.(int)
	if !ok {
		forbidden(w)
		return
	}

	user, err := s.app.GetAccount(r.Context(), id)
	if err != nil {
		writeResponseError(w, err, s.logger, http.StatusBadGateway)
		return
	}

	writeResponseSuccess(w, user, s.logger)
}

func (s *handler) Deposit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeResponseError(w, err, s.logger, http.StatusBadGateway)
		return
	}

	var data DepositRequest
	err = json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		writeResponseError(w, err, s.logger, http.StatusBadGateway)
		return
	}

	id, err := s.app.Deposit(r.Context(), accountID, data.Amount)
	if err != nil {
		writeResponseError(w, err, s.logger, http.StatusBadGateway)
		return
	}

	writeResponseSuccess(w, id, s.logger)
}

func (s *handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeResponseError(w, err, s.logger, http.StatusBadGateway)
		return
	}

	var data WithdrawRequest
	err = json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		writeResponseError(w, err, s.logger, http.StatusBadGateway)
		return
	}

	id, err := s.app.Withdraw(r.Context(), accountID, data.OrderID, data.Amount)
	if err != nil {
		writeResponseError(w, err, s.logger, http.StatusPaymentRequired)
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

func writeResponseError(w http.ResponseWriter, err error, logger Logger, statusCode int) {
	w.WriteHeader(statusCode)
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
