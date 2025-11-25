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
