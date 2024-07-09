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
	"github.com/zaytcevcom/msa/internal/logger"
	"github.com/zaytcevcom/msa/internal/rabbitmq"
	internalhttp "github.com/zaytcevcom/msa/internal/server/http"
	"github.com/zaytcevcom/msa/internal/server/metrics"
	sqlstorageuser "github.com/zaytcevcom/msa/internal/storage/user/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/demo/config.json", "Path to configuration file")
}

func main() {
	flag.Parse()

	config, err := LoadConfig(configFile)
	if err != nil {
		fmt.Println("Error loading config: ", err)
		return
	}

	port := 8000
	portMetrics := 9093

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg := logger.New(config.Logger.Level, nil)

	storage := sqlstorageuser.New(config.Postgres.Dsn)
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

	go func() {
		_ = metrics.Listen("", portMetrics)
	}()

	broker, err := rabbitmq.NewRabbitMQ(logg, config.Rabbit.URI, config.Rabbit.Exchange, config.Rabbit.Queue)
	if err != nil {
		fmt.Println("cannot connect to rabbit", err)
		return
	}

	demoApp := app.New(logg, storage, broker)

	server := internalhttp.New(logg, demoApp, "", port)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info(fmt.Sprintf("Demo service listening on port: %d, metrics port: %d", port, portMetrics))

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
