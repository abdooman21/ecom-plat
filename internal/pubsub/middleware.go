package pubsub

import (
	"log"
	"time"
)

// RetryMiddleware wraps a handler and retries it 'maxRetries' times
// before finally giving up and returning Discard.
func RetryMiddleware[T any](maxRetries int, delay time.Duration, handler func(*T) AckType) func(*T) AckType {
	return func(msg *T) AckType {
		for attempt := 0; attempt < maxRetries; attempt++ {
			result := handler(msg)
			if result == Ack {
				return Ack
			}
			if attempt < maxRetries-1 {
				log.Printf("⚠️  Handler failed (attempt %d/%d), retrying in %v...", attempt+1, maxRetries, delay)
				time.Sleep(delay)
			}
		}
		log.Printf("❌ Handler failed after %d attempts, discarding message", maxRetries)
		return Discard
	}
}
