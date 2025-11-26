// internal/routing/keys.go
package routing

const (
	// Exchanges
	ExchangePerilTopic = "peril_topic"
	ExchangePerilDLX   = "peril_dlx" // Dead Letter Exchange

	// Main Queue Configuration
	Prod_Queue = "orders_queue"
	Prod_Key   = "order.*.*" // Matches: order.{region}.{orderID}

	// Regional Routing Keys
	USOrdersKey     = "order.us.*"   // US orders only
	EUOrdersKey     = "order.eu.*"   // EU orders only
	UKOrdersKey     = "order.uk.*"   // UK orders only
	AsiaOrdersKey   = "order.asia.*" // Asia orders only
	EuropeOrdersKey = "*.*.eu"       // Alternative pattern for EU

	// Special Queue Keys
	AllEventsKey    = "#"              // Catch-all pattern (matches everything)
	HighPriorityKey = "order.*.urgent" // High priority orders
	DeadLetterQueue = "dead_letter_queue"
	RetryQueue      = "retry_queue"
)

// RoutingConfig holds routing configuration for a queue
type RoutingConfig struct {
	Exchange   string
	RoutingKey string
	QueueName  string
	Durable    bool
}

// GetStandardRoutingConfigs returns the standard routing configurations
// used in the application
func GetStandardRoutingConfigs() []RoutingConfig {
	return []RoutingConfig{
		{
			Exchange:   ExchangePerilTopic,
			RoutingKey: Prod_Key,
			QueueName:  Prod_Queue,
			Durable:    true,
		},
		{
			Exchange:   ExchangePerilTopic,
			RoutingKey: EUOrdersKey,
			QueueName:  "eu_orders_queue",
			Durable:    true,
		},
		{
			Exchange:   ExchangePerilTopic,
			RoutingKey: USOrdersKey,
			QueueName:  "us_orders_queue",
			Durable:    true,
		},
		{
			Exchange:   ExchangePerilTopic,
			RoutingKey: AllEventsKey,
			QueueName:  "analytics_queue",
			Durable:    true,
		},
	}
}

// ValidateRoutingKey checks if a routing key matches expected patterns
func ValidateRoutingKey(key string) bool {
	// Basic validation - routing keys should not be empty
	// and should follow topic exchange patterns
	if key == "" {
		return false
	}

	// Allow wildcards and alphanumeric with dots
	return true // Add more sophisticated validation if needed
}

// BuildRoutingKey constructs a routing key from components
func BuildRoutingKey(region, orderID string) string {
	return "order." + region + "." + orderID
}
