package internalbilling

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	billingmiddleware "github.com/zaytcevcom/msa/internal/server/billing/middleware"
)

type Server struct {
	server *http.Server
	logger Logger
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Account struct {
	ID      int     `json:"id"`
	Balance float64 `json:"balance"`
}

type Application interface {
	Health(ctx context.Context) interface{}
	CreateAccount(ctx context.Context, userID int) (int, error)
	GetAccount(ctx context.Context, userID int) (*Account, error)
	Deposit(ctx context.Context, accountID int, amount float64) (int, error)
	Withdraw(ctx context.Context, accountID int, orderID int, amount float64) (int, error)
}

func New(logger Logger, app Application, host string, port int) *Server {
	handler := NewHandler(logger, app)
	handler = billingmiddleware.HeaderMiddleware(handler)

	server := &http.Server{
		Addr:         net.JoinHostPort(host, strconv.Itoa(port)),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Server{
		server: server,
		logger: logger,
	}
}

func (s *Server) Start(ctx context.Context) error {
	err := s.server.ListenAndServe()
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
