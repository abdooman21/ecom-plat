// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RabbitMQ RabbitMQConfig
	App      AppConfig
}

type RabbitMQConfig struct {
	URL              string
	ReconnectDelay   time.Duration
	MaxReconnect     int
	PrefetchCount    int
	HeartbeatSeconds int
}

type AppConfig struct {
	Environment             string
	ServiceName             string
	LogLevel                string
	GracefulShutdownTimeout time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		RabbitMQ: RabbitMQConfig{
			URL:              getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			ReconnectDelay:   getDurationEnv("RABBITMQ_RECONNECT_DELAY", 5*time.Second),
			MaxReconnect:     getIntEnv("RABBITMQ_MAX_RECONNECT", 10),
			PrefetchCount:    getIntEnv("RABBITMQ_PREFETCH_COUNT", 10),
			HeartbeatSeconds: getIntEnv("RABBITMQ_HEARTBEAT_SECONDS", 10),
		},
		App: AppConfig{
			Environment:             getEnv("APP_ENV", "development"),
			ServiceName:             getEnv("SERVICE_NAME", "ecom-platform"),
			LogLevel:                getEnv("LOG_LEVEL", "info"),
			GracefulShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 30*time.Second),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.RabbitMQ.URL == "" {
		return fmt.Errorf("RABBITMQ_URL is required")
	}
	if c.RabbitMQ.PrefetchCount < 1 {
		return fmt.Errorf("RABBITMQ_PREFETCH_COUNT must be at least 1")
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
