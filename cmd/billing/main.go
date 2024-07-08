package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zaytcevcom/msa/internal/billing"
	"github.com/zaytcevcom/msa/internal/logger"
	internalbilling "github.com/zaytcevcom/msa/internal/server/billing"
	sqlstoragebilling "github.com/zaytcevcom/msa/internal/storage/billing/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/billing/config.json", "Path to configuration file")
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

	storage := sqlstoragebilling.New(config.Postgres.Dsn)
	if err = storage.Connect(ctx); err != nil {
		fmt.Println("cannot connect to storage: %w", err)
		return
	}

	defer func(storage billing.Storage, ctx context.Context) {
		err := storage.Close(ctx)
		if err != nil {
			fmt.Println("Cannot close storage connection", err)
		}
	}(storage, ctx)

	billingApp := billing.New(logg, storage)

	port := 8002
	server := internalbilling.New(logg, billingApp, "", port)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info(fmt.Sprintf("Billing service listening on port: %d", port))

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
