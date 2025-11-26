MAKE A peril_dlx , FANOUT, DURABLE : peril_dlq , durable


# Run RabbitMQ with the Management Plugin (UI)
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# E-Commerce Platform - RabbitMQ Producer/Consumer

A production-ready message queue system built with Go and RabbitMQ, featuring automatic reconnection, graceful shutdown, metrics collection, and comprehensive error handling.

## 🚀 Features

- **Auto-Reconnect**: Automatic connection recovery with configurable retry logic
- **Graceful Shutdown**: Clean shutdown with configurable timeout
- **Metrics Collection**: Built-in metrics for monitoring orders, durations, and values
- **Topic-based Routing**: Flexible message routing using RabbitMQ topic exchanges
- **Retry Middleware**: Configurable retry logic for failed message processing
- **Dead Letter Queue**: Automatic handling of failed messages
- **Docker Support**: Complete Docker Compose setup for easy deployment
- **Configuration Management**: Environment-based configuration with validation
- **Type Safety**: Generic message handlers with type-safe unmarshalling
- **Production Logging**: Structured logging with emojis for better readability

## 📁 Project Structure

```
ecom-plat/
├── cmd/
│   ├── producer/
│   │   └── main.go          # Producer application
│   └── consumer/
│       └── main.go          # Consumer application
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── handlers/
│   │   └── order_handler.go # Business logic handlers
│   ├── metrics/
│   │   └── collector.go     # Metrics collection
│   ├── pubsub/
│   │   └── pubsub.go        # RabbitMQ wrapper
│   └── routing/
│       └── keys.go          # Routing configuration
├── docker-compose.yml       # Docker orchestration
├── Dockerfile              # Multi-stage build
├── Makefile               # Development commands
├── go.mod
└── README.md
```

## 🔧 Prerequisites

- Go 1.22+
- Docker & Docker Compose (for containerized setup)
- RabbitMQ 3.12+ (if running locally)

## 📦 Installation

1. **Clone the repository**
```bash
git clone https://github.com/abdooman21/ecom-plat.git
cd ecom-plat
```

2. **Install dependencies**
```bash
make install
```

3. **Create environment configuration**
```bash
make env-example
cp .env.example .env
# Edit .env with your configuration
```

## 🚀 Quick Start

### Using Docker (Recommended)

Start all services with a single command:
```bash
make docker-up
```

This starts:
- RabbitMQ server with management UI
- Producer service
- 2 Consumer instances (scalable)

Access RabbitMQ Management UI at http://localhost:15672 (admin/admin123)

### Local Development

1. **Start RabbitMQ**
```bash
make rabbitmq-start
```

2. **Run Consumer** (in one terminal)
```bash
make run-consumer
```

3. **Run Producer** (in another terminal)
```bash
make run-producer
```

## 🎯 Usage Examples

### Publishing Messages

The producer automatically publishes orders every 2 seconds with different regions:

```go
order := Order{
    ID:        uuid.New().String(),
    Item:      "MacBook Pro",
    Price:     1999.99,
    Region:    "us",
    Timestamp: time.Now().UTC(),
}

// Routing key: order.{region}.{id}
routingKey := fmt.Sprintf("order.%s.%s", order.Region, order.ID)
pubsub.PubJSONwithCTX(ctx, ch, routing.ExchangePerilTopic, routingKey, order)
```

### Consuming Messages

The consumer has multiple subscriptions with different routing patterns:

```go
// Main orders (all regions)
pubsub.Subscribe(conn, exchange, "orders_queue", "order.*.*", pubsub.Durable, handler, unmarshaller)

// EU orders only
pubsub.Subscribe(conn, exchange, "eu_orders_queue", "order.eu.*", pubsub.Durable, handler, unmarshaller)

// Analytics (catch-all)
pubsub.Subscribe(conn, exchange, "analytics_queue", "#", pubsub.Durable, handler, unmarshaller)
```

## ⚙️ Configuration

Environment variables (see `.env.example`):

| Variable | Default | Description |
|----------|---------|-------------|
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection string |
| `APP_ENV` | `development` | Application environment |
| `SERVICE_NAME` | `ecom-platform` | Service name for logging |
| `LOG_LEVEL` | `info` | Logging level |
| `RABBITMQ_RECONNECT_DELAY` | `5s` | Delay between reconnect attempts |
| `RABBITMQ_MAX_RECONNECT` | `10` | Maximum reconnection attempts |
| `RABBITMQ_PREFETCH_COUNT` | `10` | Number of unacked messages per consumer |
| `SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout |

## 📊 Monitoring & Metrics

The system collects and logs metrics every 30 seconds:

- **Counters**: orders_processed, orders_failed, payment_retries, etc.
- **Durations**: order_processing_time (average and count)
- **Values**: order_value, sale_amount (total and average)

Example output:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Metrics Report [order-consumer]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📈 Counters:
   • orders_processed: 142
   • eu_orders_processed: 34
   • analytics_events: 142
⏱️  Durations:
   • order_processing_time: avg=203ms, count=142
💰 Values:
   • order_value: total=184,293.58, avg=1,297.84, count=142
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## 🔄 Routing Patterns

The system uses topic exchanges for flexible routing:

| Pattern | Description | Example |
|---------|-------------|---------|
| `order.*.*` | All orders | `order.us.abc123` |
| `order.eu.*` | EU orders only | `order.eu.xyz789` |
| `order.us.*` | US orders only | `order.us.abc123` |
| `#` | Catch-all (analytics) | Matches everything |
| `*.*.eu` | Alternative EU pattern | `order.any.eu` |

## 🛠️ Make Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make build` | Build producer and consumer binaries |
| `make test` | Run tests with coverage |
| `make lint` | Run linter |
| `make docker-up` | Start all services |
| `make docker-down` | Stop all services |
| `make docker-logs` | View logs |
| `make scale-consumers N=5` | Scale consumer instances |
| `make dev` | Start development environment |
| `make clean` | Clean build artifacts |

## 🐳 Docker Commands

```bash
# Start services
docker-compose up -d

# Scale consumers
docker-compose up -d --scale consumer=5

# View logs
docker-compose logs -f consumer

# Stop services
docker-compose down

# Remove volumes
docker-compose down -v
```

## 🔒 Production Considerations

1. **Security**
   - Use strong RabbitMQ credentials
   - Enable TLS for production connections
   - Implement authentication/authorization

2. **Scalability**
   - Scale consumers horizontally: `make scale-consumers N=10`
   - Adjust prefetch count based on message processing time
   - Use multiple queues for different priorities

3. **Reliability**
   - Enable message persistence (`DeliveryMode: amqp.Persistent`)
   - Configure dead letter exchanges for failed messages
   - Set up monitoring and alerting

4. **Performance**
   - Tune prefetch count for optimal throughput
   - Use connection pooling for producers
   - Monitor queue lengths and processing times

## 🧪 Testing

```bash
# Run all tests
make test

# Run short tests
make test-short

# View coverage
open coverage.html
```

## 📝 Adding New Message Types

1. Define your struct in the handler:
```go
type CustomMessage struct {
    ID   string `json:"id"`
    Data string `json:"data"`
}
```

2. Create a handler:
```go
func (h *Handler) ProcessCustom(msg *CustomMessage) pubsub.AckType {
    // Your logic here
    return pubsub.Ack
}
```

3. Subscribe in consumer:
```go
pubsub.Subscribe(
    conn,
    exchange,
    "custom_queue",
    "custom.key",
    pubsub.Durable,
    handler.ProcessCustom,
    pubsub.JSONUnmarshaller[CustomMessage],
)
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

- [RabbitMQ](https://www.rabbitmq.com/) for the excellent message broker
- [amqp091-go](https://github.com/rabbitmq/amqp091-go) for the Go client library

## 📞 Support

For issues and questions:
- Create an issue in the repository
- Check RabbitMQ documentation: https://www.rabbitmq.com/documentation.html