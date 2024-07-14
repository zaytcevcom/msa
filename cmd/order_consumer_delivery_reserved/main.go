package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/zaytcevcom/msa/internal/logger"
	consumer "github.com/zaytcevcom/msa/internal/order_consumer_delivery_reserved"
	"github.com/zaytcevcom/msa/internal/rabbitmq"
	sqlstorageorder "github.com/zaytcevcom/msa/internal/storage/order/sql"
)

var configFile string

func init() {
	flag.StringVar(
		&configFile,
		"config",
		"configs/order_consumer_delivery_reserved/config.toml",
		"Path to configuration file",
	)
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

	storage := sqlstorageorder.New(config.Postgres.Dsn)

	if err := storage.Connect(ctx); err != nil {
		fmt.Println("cannot connect to psql: ", err)
		return
	}

	defer func(storage *sqlstorageorder.Storage, _ context.Context) {
		err := storage.Close(ctx)
		if err != nil {
			fmt.Println("Cannot close connection", err)
		}
	}(storage, ctx)

	brokerConsumer, err := rabbitmq.NewRabbitMQ(
		logg,
		config.RabbitConsumer.URI,
		config.RabbitConsumer.Exchange,
		config.RabbitConsumer.Queue,
	)
	if err != nil {
		fmt.Println("cannot connect to consumer", err)
		return
	}

	s := consumer.New(logg, storage, brokerConsumer)

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	if err := s.Start(); err != nil {
		logg.Error("failed to start: " + err.Error())
	}
}
