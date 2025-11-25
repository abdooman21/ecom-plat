package main

import (
	"log"

	"github.com/abdooman21/ecom-plat/internal/pubsub"
	"github.com/abdooman21/ecom-plat/internal/routing"
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

	failOnError(err, "Failed to declare a queue")
	pubsub.Subscribe(conn, routing.ExchangePerilTopic, routing.Prod_Queue, routing.Prod_Key, pubsub.Durable, func(msg *Order) pubsub.AckType {
		log.Printf("Received Order: %+v", msg)
		return pubsub.Ack
	}, pubsub.JSONUnmarshaller[Order])

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
