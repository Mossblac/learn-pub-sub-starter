package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueDurable = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	RabChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("error creating channel: %v", err)
	}

	Am_queue, err := RabChan.QueueDeclare(
		queueName,
		queueType == SimpleQueueDurable,
		queueType != SimpleQueueDurable,
		queueType != SimpleQueueDurable,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("error creating queue: %v", err)
	}

	err = RabChan.QueueBind(
		Am_queue.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not bind queue: %v", err)
	}
	return RabChan, Am_queue, nil
}

//durable = durable- true
//autoDelete = transient - true
//exclusive = transient - true
//noWait = false
//args = nil
