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
	ID    string  `json:"id"`
	Item  string  `json:"item"`
	Price float64 `json:"price"`
}

func main() {
	// Get RabbitMQ URL from environment or use default
	url := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	log.Println("🚀 Starting Consumer Service")

	// Connect to RabbitMQ
	conn := pubsub.Connect_RabbitMQ(url)
	defer func() {
		conn.Close()
		log.Println("🔌 Connection closed to RabbitMQ")
	}()

	// Open channel for setup
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// Declare exchange
	err = ch.ExchangeDeclare(
		routing.ExchangePerilTopic,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	// ========================================
	// CONSUMER 1: Main Orders Queue
	// ========================================
	log.Printf("📬 [1] Subscribing to: %s (key: %s)", routing.Prod_Queue, routing.Prod_Key)

	orderHandler := pubsub.RetryMiddleware(3, 1*time.Second, func(msg *Order) pubsub.AckType {
		log.Printf("📦 Order Received: %s | %s | $%.2f", msg.ID, msg.Item, msg.Price)

		// Your business logic here
		if err := processOrder(msg); err != nil {
			log.Printf("⚠️  Error processing order: %v", err)
			return pubsub.Requeue
		}

		return pubsub.Ack
	})

	err = pubsub.Subscribe(
		conn,
		routing.ExchangePerilTopic,
		routing.Prod_Queue,
		routing.Prod_Key,
		pubsub.Durable,
		orderHandler,
		pubsub.JSONUnmarshaller[Order],
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to main queue: %v", err)
	}

	// ========================================
	// CONSUMER 2: EU Orders Queue
	// ========================================
	log.Printf("🇪🇺 [2] Subscribing to: eu_orders_queue (key: %s)", routing.EuropeOrdersKey)

	err = pubsub.Subscribe(
		conn,
		routing.ExchangePerilTopic,
		"eu_orders_queue",
		routing.EuropeOrdersKey,
		pubsub.Durable,
		func(msg *Order) pubsub.AckType {
			log.Printf("🇪🇺 EU Order: %s | %s | €%.2f", msg.ID, msg.Item, msg.Price)
			// EU-specific processing
			return pubsub.Ack
		},
		pubsub.JSONUnmarshaller[Order],
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to EU queue: %v", err)
	}

	// ========================================
	// CONSUMER 3: Analytics Queue (Catch-all)
	// ========================================
	log.Printf("📊 [3] Subscribing to: analytics_queue (key: %s)", routing.AllEventsKey)

	err = pubsub.Subscribe(
		conn,
		routing.ExchangePerilTopic,
		"analytics_queue",
		routing.AllEventsKey,
		pubsub.Durable,
		func(msg *Order) pubsub.AckType {
			log.Printf("📊 Analytics: %s - $%.2f", msg.Item, msg.Price)
			// Save to analytics database, update dashboards, etc.
			return pubsub.Ack
		},
		pubsub.JSONUnmarshaller[Order],
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to analytics queue: %v", err)
	}

	log.Println("✅ All consumers ready and listening")
	log.Println("⏳ Press CTRL+C to exit...")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("🛑 Shutting down gracefully...")
	time.Sleep(2 * time.Second) // Allow time for in-flight messages
}

// processOrder simulates order processing logic
func processOrder(order *Order) error {
	// Validate order
	if order.Price <= 0 {
		log.Printf("❌ Invalid price for order %s", order.ID)
		return nil // Don't retry invalid data
	}

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Your actual business logic:
	// - Charge payment
	// - Update inventory
	// - Send confirmation email
	// - etc.

	log.Printf("✅ Order %s processed successfully", order.ID)
	return nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
