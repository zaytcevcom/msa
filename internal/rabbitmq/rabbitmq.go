package rabbitmq

import (
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	logger   Logger
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    amqp.Queue
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

func NewRabbitMQ(logger Logger, uri string, exchange string, queue string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err = ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if err = ch.QueueBind(
		queue,
		queue,
		exchange,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	return &RabbitMQ{
		logger:   logger,
		conn:     conn,
		exchange: exchange,
		channel:  ch,
		queue:    q,
	}, nil
}

func (r *RabbitMQ) Publish(body string) error {
	return r.channel.Publish(
		r.exchange,
		r.queue.Name,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Persistent,
			Priority:        0,
		})
}

func (r *RabbitMQ) Subscribe(handler func(body []byte) error) error {
	messages, err := r.channel.Consume(
		r.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for d := range messages {
		err := handler(d.Body)
		if err != nil {
			r.logger.Error(err.Error())
		}
	}

	return nil
}

func (r *RabbitMQ) Close() error {
	return r.conn.Close()
}
