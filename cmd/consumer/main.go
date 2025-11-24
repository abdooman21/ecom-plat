package main

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Order struct {
	ID   string `json:"id"`
	Item string `json:"item"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"orders_queue", // Must match the Producer's queue name
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	// 1. Register a Consumer
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer tag (empty = auto generated)
		true,   // auto-ack (We will set this to FALSE later for safety, but TRUE is easier for now)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	// 2. Read messages forever using a channel
	var forever chan struct{}

	go func() {
		for d := range msgs {
			var order Order
			json.Unmarshal(d.Body, &order)

			log.Printf(" [📦] Inventory Service: Processing Order %s", order.ID)
			log.Printf("      -> Decrementing stock for: %s", order.Item)
			// Simulate DB work
		}
	}()

	log.Printf(" [*] Waiting for orders. To exit press CTRL+C")
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
