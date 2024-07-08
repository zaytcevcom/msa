package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/zaytcevcom/msa/internal/logger"
	"github.com/zaytcevcom/msa/internal/notification_sender"
	"github.com/zaytcevcom/msa/internal/rabbitmq"
	sqlstoragenotification "github.com/zaytcevcom/msa/internal/storage/notification/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/notification_sender/config.toml", "Path to configuration file")
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

	storage := sqlstoragenotification.New(config.Postgres.Dsn)

	if err := storage.Connect(ctx); err != nil {
		fmt.Println("cannot connect to psql: ", err)
		return
	}

	defer func(storage *sqlstoragenotification.Storage, _ context.Context) {
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

	s := notificationsender.New(logg, storage, broker)

	go func() {
		<-ctx.Done()

		if err := s.Stop(); err != nil {
			fmt.Println("Cannot close rabbit connection", err)
		}
	}()

	if err := s.Start(); err != nil {
		logg.Error("failed to start grpc server: " + err.Error())
	}
}
