package app

import "context"

type App struct {
	logger Logger
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

func New(logger Logger) *App {
	return &App{
		logger: logger,
	}
}

func (a App) Health(_ context.Context) string {
	return "OK"
}
