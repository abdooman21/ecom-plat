package pubsub

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable = iota
	Transient
)

func Connect_RabbitMQ(url string) (conn *amqp.Connection) {
	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Printf("failed at connection, ")
		// exit
		log.Fatal(err)
	}
	log.Println("Connected to RabbitMQ successfully")
	return

}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to open channel: %w", err)
	}

	transient := queueType == Transient
	durable := Durable == queueType

	table := amqp.Table{"x-dead-letter-exchange": "peril_dlx"}

	qu, err := ch.QueueDeclare(queueName, durable, transient, transient, false, table)

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to open queue: %w", err)

	}

	if err := ch.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to bind queue: %w", err)

	}

}
