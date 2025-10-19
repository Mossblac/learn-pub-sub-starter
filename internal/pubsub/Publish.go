package pubsub

import amqp "github.com/rabbitmq/amqp091-go"

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	return nil
}
