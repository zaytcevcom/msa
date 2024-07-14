package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	deliveryconsumer "github.com/zaytcevcom/msa/internal/delivery_consumer"
	"github.com/zaytcevcom/msa/internal/logger"
	"github.com/zaytcevcom/msa/internal/rabbitmq"
	sqlstoragedelivery "github.com/zaytcevcom/msa/internal/storage/delivery/sql"
)

var configFile string

func init() {
	flag.StringVar(
		&configFile,
		"config",
		"configs/delivery_consumer_stock_reserved/config.toml",
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

	storage := sqlstoragedelivery.New(config.Postgres.Dsn)

	if err := storage.Connect(ctx); err != nil {
		fmt.Println("cannot connect to psql: ", err)
		return
	}

	defer func(storage *sqlstoragedelivery.Storage, _ context.Context) {
		err := storage.Close(ctx)
		if err != nil {
			fmt.Println("Cannot close connection", err)
		}
	}(storage, ctx)

	consumer, err := rabbitmq.NewRabbitMQ(
		logg,
		config.RabbitConsumer.URI,
		config.RabbitConsumer.Exchange,
		config.RabbitConsumer.Queue,
	)
	if err != nil {
		fmt.Println("cannot connect to consumer", err)
		return
	}

	producerSuccess, err := rabbitmq.NewRabbitMQ(
		logg,
		config.RabbitProducerSuccess.URI,
		config.RabbitProducerSuccess.Exchange,
		config.RabbitProducerSuccess.Queue,
	)
	if err != nil {
		fmt.Println("cannot connect to producerSuccess", err)
		return
	}

	producerReject, err := rabbitmq.NewRabbitMQ(
		logg,
		config.RabbitProducerReject.URI,
		config.RabbitProducerReject.Exchange,
		config.RabbitProducerReject.Queue,
	)
	if err != nil {
		fmt.Println("cannot connect to producerReject", err)
		return
	}

	s := deliveryconsumer.New(logg, storage, consumer, producerSuccess, producerReject)

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	if err := s.Start(); err != nil {
		logg.Error("failed to start: " + err.Error())
	}
}
