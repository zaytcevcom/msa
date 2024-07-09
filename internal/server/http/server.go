package internalhttp

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/zaytcevcom/msa/internal/server/http/middleware"
	"github.com/zaytcevcom/msa/internal/storage/user"
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

type Application interface {
	Health(ctx context.Context) interface{}
	GetByID(ctx context.Context, id int) (*storageuser.Entity, error)
	Create(
		ctx context.Context,
		username string,
		password string,
		firstName string,
		lastName string,
		email string,
		phone string,
	) (int, error)
	Update(ctx context.Context, id int, user storageuser.Entity) error
	Delete(ctx context.Context, id int) error
}

func New(logger Logger, app Application, host string, port int) *Server {
	handler := NewHandler(logger, app)
	handler = middleware.PrometheusMiddleware(handler)
	handler = middleware.HeaderMiddleware(handler)

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
