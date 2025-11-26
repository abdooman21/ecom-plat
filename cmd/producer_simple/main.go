// cmd/producer_simple/main.go
// A minimal producer without extra dependencies
// Perfect for getting started quickly

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
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

	log.Println("🚀 Starting Producer Service")

	// Connect to RabbitMQ
	conn := pubsub.Connect_RabbitMQ(url)
	defer func() {
		conn.Close()
		log.Println("🔌 Connection closed to RabbitMQ")
	}()

	// Open channel
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

	log.Println("✅ Producer ready")
	log.Println("📤 Publishing orders every 2 seconds...")
	log.Println("💡 Routing patterns:")
	log.Println("   order.us.*   → US orders")
	log.Println("   order.eu.*   → EU orders")
	log.Println("   order.uk.*   → UK orders")
	log.Println("   order.asia.* → Asia orders")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Publish timer
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	rand.Seed(time.Now().UnixNano())
	count := 0

	for {
		select {
		case <-ticker.C:
			// Create random order
			order := generateRandomOrder(count)
			count++

			// Build routing key
			region := pickRandomRegion()
			routingKey := fmt.Sprintf("order.%s.%s", region, order.ID)

			// Publish
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := pubsub.PubJSONwithCTX(ctx, ch, routing.ExchangePerilTopic, routingKey, order)
			cancel()

			if err != nil {
				log.Printf("❌ Failed to publish: %v", err)
			} else {
				log.Printf("📤 [%d] Published: %s | %s | $%.2f → %s",
					count, order.ID, order.Item, order.Price, routingKey)
			}

		case <-sigChan:
			log.Println("🛑 Shutdown requested")
			log.Printf("📊 Total published: %d orders", count)
			return
		}
	}
}

// generateRandomOrder creates a mock order
func generateRandomOrder(num int) Order {
	products := []struct {
		name  string
		price float64
	}{
		{"MacBook Pro 16\"", 2499.99},
		{"iPhone 15 Pro Max", 1199.99},
		{"AirPods Pro 2", 249.99},
		{"iPad Pro 12.9\"", 1099.99},
		{"Apple Watch Ultra 2", 799.99},
		{"Magic Keyboard", 349.99},
		{"Studio Display", 1599.99},
		{"HomePod", 299.99},
		{"Apple TV 4K", 179.99},
		{"AirTag 4-pack", 99.99},
	}

	product := products[rand.Intn(len(products))]

	return Order{
		ID:    fmt.Sprintf("ORD-%d", 1000+num),
		Item:  product.name,
		Price: product.price,
	}
}

// pickRandomRegion selects a random region for routing
func pickRandomRegion() string {
	regions := []string{"us", "eu", "uk", "asia"}
	return regions[rand.Intn(len(regions))]
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
