package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
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

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {

	var buf bytes.Buffer
	gobData := gob.NewEncoder(&buf)
	err := gobData.Encode(val)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	if err != nil {
		log.Fatalf("basic.publish: %v", err)
	}

	return nil
}

func CreateGameLog(time time.Time, msg, Uname string) (Gl routing.GameLog) {
	Gl = routing.GameLog{
		CurrentTime: time,
		Message:     msg,
		Username:    Uname,
	}
	return Gl
}
