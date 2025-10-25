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
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	MoveChannel, Movequeue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		"army_moves."+username,
		"army_moves.*",
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatalf("could not subscribe to moves: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", Movequeue.Name)

	defer conn.Close()
	fmt.Println("Connected to RabbitMQ")
	fmt.Println("Starting Peril server...")

	gamestate := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		"pause."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gamestate),
	)

	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		"army_moves.*",
		pubsub.SimpleQueueTransient,
		handlerMove(MoveChannel, gamestate),
	)

	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"war",
		routing.WarRecognitionsPrefix+".#",
		pubsub.SimpleQueueDurable,
		handlerWar(gamestate),
	)

	for {
		words := gamelogic.GetInput()
		switch words[0] {
		case "spawn":
			err := gamestate.CommandSpawn(words)
			if err != nil {
				fmt.Printf("unable to spawn unit: %v", err)
			}
		case "move":
			MoveResult, err := gamestate.CommandMove(words)
			if err != nil {
				fmt.Printf("unable to move unit: %v", err)
			}
			err = pubsub.PublishJSON(MoveChannel, routing.ExchangePerilTopic, "army_moves."+username, MoveResult)
			if err != nil {
				fmt.Printf("unable to publish move: %v", err)
			}
			fmt.Println("Move Published")
		case "status":
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet")
		case "quit":
			gamelogic.PrintQuit()
			fmt.Printf("\nShutting Down RabbitMQ connection\n Ending Program...\n")
			conn.Close()
			return
		default:
			fmt.Println("Unkown Command")
		}
	}
}
