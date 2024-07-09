package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zaytcevcom/msa/internal/logger"
	"github.com/zaytcevcom/msa/internal/order"
	"github.com/zaytcevcom/msa/internal/rabbitmq"
	internalredis "github.com/zaytcevcom/msa/internal/redis"
	internalorder "github.com/zaytcevcom/msa/internal/server/order"
	sqlstorageorder "github.com/zaytcevcom/msa/internal/storage/order/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/order/config.json", "Path to configuration file")
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

	storage := sqlstorageorder.New(config.Postgres.Dsn)
	if err = storage.Connect(ctx); err != nil {
		fmt.Println("cannot connect to storage: %w", err)
		return
	}

	defer func(storage order.Storage, ctx context.Context) {
		err := storage.Close(ctx)
		if err != nil {
			fmt.Println("Cannot close storage connection", err)
		}
	}(storage, ctx)

	broker, err := rabbitmq.NewRabbitMQ(logg, config.Rabbit.URI, config.Rabbit.Exchange, config.Rabbit.Queue)
	if err != nil {
		fmt.Println("cannot connect to rabbit", err)
		return
	}

	redis := internalredis.NewRedis(logg, config.Redis.Host, config.Redis.Port, config.Redis.Password)
	if redis == nil {
		fmt.Println("cannot connect to redis")
		return
	}

	orderApp := order.New(logg, storage, broker, redis)

	port := 8004
	server := internalorder.New(logg, orderApp, "", port)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info(fmt.Sprintf("Order service listening on port: %d", port))

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
