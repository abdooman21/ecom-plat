package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable = iota
	Transient
)

type AckType int

const (
	Ack AckType = iota
	Requeue
	Discard
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
	return ch, qu, nil
}

func Subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	QueueType SimpleQueueType,
	handler func(*T) AckType,
	unmarshaller func([]byte) (*T, error),
) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, QueueType)
	if err != nil {
		return fmt.Errorf("at declaring and binding: %w", err)
	}
	err = ch.Qos(10, 0, false)
	if err != nil {
		return fmt.Errorf("failed prefetch limit: %w", err)
	}
	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}
	go func() {
		for d := range msgs {
			msg, err := unmarshaller(d.Body)
			if err != nil {
				log.Printf("failed to decode message: %v", err)
				d.Nack(false, false) // discard
				continue
			}
			ack := handler(msg)
			switch ack {
			case Ack:
				d.Ack(false)
				log.Println(" Hancdler Ack meesage ")
			case Requeue:
				d.Nack(false, true)
				log.Println(" Hancdler requeue meesage ")
			case Discard:
				d.Nack(false, false)
				log.Println(" Hancdler discard meesage ")
			}
		}
	}()

	return nil
}

func PubGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(val); err != nil {
		return err
	}
	return ch.PublishWithContext(context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Timestamp:   time.Now().UTC(),
			Body:        buf.Bytes(),
		},
	)

}
func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	body, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Timestamp:   time.Now().UTC(),
			Body:        body,
		},
	)
}
func PubJSONwithCTX[T any](ctx context.Context, ch *amqp.Channel, exchange, key string, val T) error {
	body, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return ch.PublishWithContext(ctx,
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Timestamp:   time.Now().UTC(),
			Body:        body,
		},
	)

}
func JSONUnmarshaller[T any](body []byte) (*T, error) {
	var msg T
	err := json.Unmarshal(body, &msg)
	return &msg, err

}
func Gobunmarshaller[T any](body []byte) (*T, error) {
	var msg T
	buf := bytes.NewBuffer(body)
	err := gob.NewDecoder(buf).Decode(&msg)
	return &msg, err
}
