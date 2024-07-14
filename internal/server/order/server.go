package internalorder

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	ordermiddleware "github.com/zaytcevcom/msa/internal/server/order/middleware"
	storageorder "github.com/zaytcevcom/msa/internal/storage/order"
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
	GetByID(ctx context.Context, id int) (*storageorder.EntityOrder, error)
	Create(ctx context.Context, userID int, productID int, sum float64, email string) (int, error)
}

func New(logger Logger, app Application, host string, port int) *Server {
	handler := NewHandler(logger, app)
	handler = ordermiddleware.HeaderMiddleware(handler)

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
