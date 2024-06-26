package internalhttp

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/zaytcevcom/msa/internal/storage"
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
	GetByID(ctx context.Context, id int) (*storage.User, error)
	Create(
		ctx context.Context,
		username string,
		firstName string,
		lastName string,
		email string,
		phone string,
	) (int, error)
	Update(ctx context.Context, id int, user storage.User) error
	Delete(ctx context.Context, id int) error
}

func New(logger Logger, app Application, host string, port int) *Server {
	server := &http.Server{
		Addr:         net.JoinHostPort(host, strconv.Itoa(port)),
		Handler:      NewHandler(logger, app),
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
