package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	connectionString := "amqp://guest:guest@localhost:5672/"
	con, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	/*RabChan, err := con.Channel() // have to create function for *amqp.Channel before it will save.
	if err != nil{
		fmt.Printf("Error with RabChan: %v", err)
	}*/

	defer con.Close()
	fmt.Println("Connected to RabbitMQ")
	fmt.Println("Starting Peril server...")

	<-sigs
	fmt.Printf("\nShutting Down RabbitMQ connection\n Ending Program...\n")
	con.Close()

}
