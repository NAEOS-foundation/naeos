package broker

import "fmt"

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
		return nil, fmt.Errorf("unsupported broker driver: %s", driver)
	}
	if err := b.Connect(config); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return b, nil
}
