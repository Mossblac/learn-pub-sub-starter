package pubsub

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
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
		for message := range sig {
			var MesT T
			err = json.Unmarshal(message.Body, &MesT)
			if err != nil {
				return err
			}
			ack := handler(MesT)
			switch ack {
			case Ack:
				err = message.Ack(false)
				if err != nil {
					return err
				}
			case NackRequeue:
				err = message.Nack(false, true)
				if err != nil {
					return err
				}
			case NackDiscard:
				err = message.Nack(false, false)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}()
	return nil
}
