package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	accountcreator "github.com/zaytcevcom/msa/internal/account_creator"
	"github.com/zaytcevcom/msa/internal/logger"
	"github.com/zaytcevcom/msa/internal/rabbitmq"
	sqlstoragebilling "github.com/zaytcevcom/msa/internal/storage/billing/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/account_creator/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	config, err := LoadSenderConfig(configFile)
	if err != nil {
		fmt.Println("Error loading config: ", err)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg := logger.New(config.Logger.Level, nil)

	storage := sqlstoragebilling.New(config.Postgres.Dsn)

	if err := storage.Connect(ctx); err != nil {
		fmt.Println("cannot connect to psql: ", err)
		return
	}

	defer func(storage *sqlstoragebilling.Storage, _ context.Context) {
		err := storage.Close(ctx)
		if err != nil {
			fmt.Println("Cannot close connection", err)
		}
	}(storage, ctx)

	broker, err := rabbitmq.NewRabbitMQ(logg, config.Rabbit.URI, config.Rabbit.Exchange, config.Rabbit.Queue)
	if err != nil {
		fmt.Println("cannot connect to rabbit", err)
		return
	}

	s := accountcreator.New(logg, storage, broker)

	go func() {
		<-ctx.Done()

		if err := s.Stop(); err != nil {
			fmt.Println("Cannot close rabbit connection", err)
		}
	}()

	if err := s.Start(); err != nil {
		logg.Error("failed to start: " + err.Error())
	}
}
