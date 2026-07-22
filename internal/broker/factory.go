package broker

import (
	"fmt"
	"log/slog"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// New creates a new broker instance by driver name.
// Supported drivers: "redis", "rabbitmq", "kafka", "nats", "memory".
// Use "mock-redis", "mock-rabbitmq", "mock-kafka" for stub adapters.
//
// Callers must check the return value for nil before use, as New returns nil
// for unrecognized driver names.
func New(driver string) Broker {
	switch driver {
	case "redis":
		return NewRealRedis()
	case "rabbitmq":
		return NewRealRabbitMQ()
	case "kafka":
		return NewRealKafka()
	case "nats":
		return NewRealNATS()
	case "memory", "inmemory":
		return NewInMemoryBroker()
	case "mock-redis":
		return NewRedis()
	case "mock-rabbitmq":
		return NewRabbitMQ()
	case "mock-kafka":
		return NewKafka()
	default:
		return nil
	}
}

// SupportedDrivers returns the list of broker driver names accepted by New and
// NewFromConfig.
func SupportedDrivers() []string {
	return []string{
		"redis",
		"rabbitmq",
		"kafka",
		"nats",
		"memory",
		"inmemory",
		"mock-redis",
		"mock-rabbitmq",
		"mock-kafka",
	}
}

// NewFromConfig creates and connects a broker by driver name and config.
func NewFromConfig(driver string, config *Config) (Broker, error) {
	b := New(driver)
	if b == nil {
		slog.Error("unsupported broker driver", "driver", driver)
		return nil, naeoserr.New(naeoserr.ErrConfig, fmt.Sprintf("unsupported broker driver: %s", driver))
	}
	if err := b.Connect(config); err != nil {
		slog.Error("broker connect failed", "driver", driver, "error", err)
		return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "connect")
	}
	return b, nil
}
