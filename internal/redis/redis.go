package internalredis

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	logger Logger
	client *redis.Client
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

func NewRedis(logger Logger, host string, port int, password string) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(host, strconv.Itoa(port)),
		Password: password,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil
	}

	return &Redis{
		logger: logger,
		client: client,
	}
}

func (c *Redis) Get(key string) (string, bool) {
	val, err := c.client.Get(context.Background(), key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false
	} else if err != nil {
		c.logger.Error(err.Error())
		return "", false
	}
	return val, true
}

func (c *Redis) Set(key string, value string) bool {
	err := c.client.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		c.logger.Error(err.Error())
		return false
	}
	return true
}
