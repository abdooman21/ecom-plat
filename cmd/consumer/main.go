package main

import (
	"log"

	"github.com/abdooman21/ecom-plat/internal/pubsub"
	"github.com/abdooman21/ecom-plat/internal/routing"
)

type Order struct {
	ID   string `json:"id"`
	Item string `json:"item"`
}

func main() {
	url := "amqp://guest:guest@localhost:5672/"
	conn := pubsub.Connect_RabbitMQ(url)

	defer func() {
		conn.Close()
		log.Println("closing connection to RabbitMQ,...  Checking if closed  \" ", conn.IsClosed(), "\"")
	}()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	defer ch.Close()

	err = pubsub.Subscribe(conn, routing.ExchangePerilTopic, routing.Prodque, routing.ProdKey+".*", pubsub.Durable, handlerProd(*Order{}), pubsub.JSONUnmarshaller[Order])
	// pubsub.DeclareAndBind(conn,)
	// q, err := ch.QueueDeclare(
	// 	"orders_queue", // Must match the Producer's queue name
	// 	true,
	// 	false,
	// 	false,
	// 	false,
	// 	nil,
	// )

	// failOnError(err, "Failed to declare a queue")

	// 1. Register a Consumer
	// msgs, err := ch.Consume(
	// 	q.Name, // queue
	// 	"",     // consumer tag (empty = auto generated)
	// 	true,   // auto-ack (We will set this to FALSE later for safety, but TRUE is easier for now)
	// 	false,  // exclusive
	// 	false,  // no-local
	// 	false,  // no-wait
	// 	nil,    // args
	// )
	// failOnError(err, "Failed to register a consumer")

	// 2. Read messages forever using a channel
	// var forever chan struct{}

	// go func() {
	// 	for d := range msgs {
	// 		var order Order
	// 		json.Unmarshal(d.Body, &order)

	// 		log.Printf(" [📦] Inventory Service: Processing Order %s", order.ID)
	// 		log.Printf("      -> Decrementing stock for: %s", order.Item)
	// 		// Simulate DB work
	// 	}
	// }()

	// log.Printf(" [*] Waiting for orders. To exit press CTRL+C")
	// <-forever

}

func handlerProd() func(order *Order) pubsub.AckType {
	panic("unimplemented")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
