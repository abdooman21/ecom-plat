package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	retryHandler := pubsub.RetryMiddleware(3, 1*time.Second, func(msg *Order) pubsub.AckType {
		log.Printf("Received Order: %+v", msg)
		return pubsub.Ack
	})
	pubsub.Subscribe(conn, routing.ExchangePerilTopic, routing.Prod_Queue, routing.Prod_Key, pubsub.Durable, retryHandler, pubsub.JSONUnmarshaller[Order])

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down consumer...")

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
