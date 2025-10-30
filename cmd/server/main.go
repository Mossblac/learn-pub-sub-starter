package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	connectionString := "amqp://guest:guest@localhost:5672/"
	con, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	RabChan, err := con.Channel()
	if err != nil {
		fmt.Printf("Error with RabChan: %v", err)
	}

	defer con.Close()
	fmt.Println("Connected to RabbitMQ")
	fmt.Println("Starting Peril server...")
	gamelogic.PrintServerHelp()

	_, queue, err := pubsub.DeclareAndBind(
		con,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.*",
		pubsub.SimpleQueueDurable,
	)
	if err != nil {
		log.Fatalf("could not subscribe to topic: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	err = pubsub.SubscribeGOB(
		con,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.#",
		pubsub.SimpleQueueDurable,
		handlerGameLog(),
	)
	if err != nil {
		defer fmt.Print("> ")
		log.Fatalf("error with SubscribeGOB: %v", err)
	}

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			fmt.Println("no input")
		} else {
			switch input[0] {
			case "pause":
				fmt.Println("sending pause message")
				err = pubsub.PublishJSON(RabChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
				if err != nil {
					fmt.Printf("Error with PublishJSON: %v", err)
				}
			case "resume":
				fmt.Println("sending resume message")
				err = pubsub.PublishJSON(RabChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
				if err != nil {
					fmt.Printf("Error with PublishJSON: %v", err)
				}
			case "quit":

				fmt.Printf("\nShutting Down RabbitMQ connection\n Ending Program...\n")
				con.Close()
				return
			default:
				fmt.Println("unknown command")
			}

		}
	}
}

func handlerGameLog() func(routing.GameLog) pubsub.Acktype {
	return func(gl routing.GameLog) pubsub.Acktype {
		defer fmt.Print("> ")
		if err := gamelogic.WriteLog(gl); err != nil {
			fmt.Println("error with WriteLog, discarded from queue")
			return pubsub.NackDiscard
		}
		return pubsub.Ack
	}
}
