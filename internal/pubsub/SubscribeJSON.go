package pubsub

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	channel, Aqueue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	sig, err := channel.Consume(Aqueue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() error {
		var MesT T
		for message := range sig {
			err = json.Unmarshal(message.Body, &MesT)
			if err != nil {
				return err
			}
			handler(MesT)
			err = message.Ack(false)
			if err != nil {
				return err
			}
		}
		return nil
	}()
	return nil
}
