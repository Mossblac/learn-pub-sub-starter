package pubsub

import (
	"bytes"
	"encoding/gob"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGOB[T any](
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
			buffer := bytes.NewBuffer(message.Body)
			dec := gob.NewDecoder(buffer)
			var g T
			err := dec.Decode(&g)
			if err != nil {
				return err
			}
			ack := handler(g)
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
