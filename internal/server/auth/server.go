package internalauth

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"
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

type Token struct {
	AccessToken string `json:"accessToken"`
	UserID      int    `json:"userId"`
}

type Application interface {
	Health(ctx context.Context) interface{}
	Auth(ctx context.Context, header http.Header) (int, error)
	Login(ctx context.Context, username string, password string) (*Token, error)
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
