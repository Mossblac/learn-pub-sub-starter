package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	jsonData, err := json.Marshal(val)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
	}
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonData,
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	if err != nil {
		log.Fatalf("basic.publish: %v", err)
	}

	return nil
}
