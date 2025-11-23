package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// 1. Define the Data Contract (What the message looks like)
type Order struct {
	ID    string  `json:"id"`
	Item  string  `json:"item"`
	Price float64 `json:"price"`
}

func main() {
	// 2. Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// 3. Declare the Queue (Idempotent: creates it if it doesn't exist)
	q, err := ch.QueueDeclare(
		"orders_queue", // name
		true,           // durable (messages survive server restart)
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// 4. Create a Mock Order
	order := Order{ID: "ORD-101", Item: "MacBook Pro", Price: 1999.99}
	body, _ := json.Marshal(order)

	// 5. Publish the Message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",     // exchange (default exchange)
		q.Name, // routing key (queue name)
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent Order: %s", body)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
