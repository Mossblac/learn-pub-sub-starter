package pubsub

import (
	"bytes"
	"encoding/gob"
	"fmt"

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
	channel.Qos(10, 0, false)

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
				message.Nack(false, true)
				fmt.Printf("error decoding %v", err)
				continue
			}
			ack := handler(g)
			switch ack {
			case Ack:
				err = message.Ack(false)
				if err != nil {
					fmt.Printf("error Ack %v", err)
				}
			case NackRequeue:
				err = message.Nack(false, true)
				if err != nil {
					fmt.Printf("error Nack %v", err)
				}
			case NackDiscard:
				err = message.Nack(false, false)
				if err != nil {
					fmt.Printf("error Nack %v", err)
				}
			}
		}
		return nil
	}()
	return nil
}
