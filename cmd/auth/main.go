package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zaytcevcom/msa/internal/app"
	"github.com/zaytcevcom/msa/internal/auth"
	"github.com/zaytcevcom/msa/internal/logger"
	internalauth "github.com/zaytcevcom/msa/internal/server/auth"
	sqlstorage "github.com/zaytcevcom/msa/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/auth/config.json", "Path to configuration file")
}

func main() {
	flag.Parse()

	config, err := LoadConfig(configFile)
	if err != nil {
		fmt.Println("Error loading config: ", err)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg := logger.New(config.Logger.Level, nil)

	storage := sqlstorage.New(config.Postgres.Dsn)
	if err = storage.Connect(ctx); err != nil {
		fmt.Println("cannot connect to storage: %w", err)
		return
	}

	defer func(storage app.Storage, ctx context.Context) {
		err := storage.Close(ctx)
		if err != nil {
			fmt.Println("Cannot close storage connection", err)
		}
	}(storage, ctx)

	app := auth.New(logg, storage)

	port := 8001
	server := internalauth.New(logg, app, "", port)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info(fmt.Sprintf("Auth service listening on port: %d", port))

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
