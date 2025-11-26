package main

import (
	"context"
	"log"
	"time"

	"github.com/abdooman21/ecom-plat/internal/pubsub"
	"github.com/abdooman21/ecom-plat/internal/routing"
)

// 1. Define the Data Contract (What the message looks like)
type Order struct {
	ID    string  `json:"id"`
	Item  string  `json:"item"`
	Price float64 `json:"price"`
}

func main() {

	url := "amqp://guest:guest@localhost:5672/"
	conn := pubsub.Connect_RabbitMQ(url)

	defer func() {
		conn.Close()
		log.Println("closing connection to RabbitMQ,...  Checking if closed  \" ", conn.IsClosed(), "\"")
	}()

	// 	"orders_queue", // name
	// 	true,           // durable (messages survive server restart)
	// 	false,          // delete when unused
	// 	false,          // exclusive
	// 	false,          // no-wait
	// 	nil,            // arguments

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open publisher channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		routing.ExchangePerilTopic, // Name
		"topic",                    // TYPE IS NOW TOPIC!
		true, false, false, false, nil,
	)

	// 4. Create a Mock Order

	order := Order{ID: "ORD-101", Item: "MacBook Pro", Price: 1999.99}

	// 5. Publish the Message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	pubsub.PubJSONwithCTX(ctx, pubCh, routing.ExchangePerilTopic, routing.Prod_Key, order)

	defer cancel()

	log.Printf(" [x] Sent Order: %v", order)
}

// func failOnError(err error, msg string) {
// 	if err != nil {
// 		log.Panicf("%s: %s", msg, err)
// 	}
// }
