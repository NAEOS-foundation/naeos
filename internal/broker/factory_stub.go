//go:build nobroker

package broker

type RealRedis struct{}
type RealRabbitMQ struct{}
type RealKafka struct{}
type RealNATS struct{}

func (r *RealRedis) Name() string                          { return "redis" }
func (r *RealRedis) Connect(config *Config) error          { return nil }
func (r *RealRedis) Close() error                          { return nil }
func (r *RealRedis) Ping() error                           { return nil }
func (r *RealRedis) Publish(channel string, msg *Message) error { return nil }
func (r *RealRedis) Subscribe(channel string, handler MessageHandler) error { return nil }
func (r *RealRedis) Unsubscribe(channel string) error      { return nil }

func (r *RealRabbitMQ) Name() string                       { return "rabbitmq" }
func (r *RealRabbitMQ) Connect(config *Config) error       { return nil }
func (r *RealRabbitMQ) Close() error                       { return nil }
func (r *RealRabbitMQ) Ping() error                        { return nil }
func (r *RealRabbitMQ) Publish(channel string, msg *Message) error { return nil }
func (r *RealRabbitMQ) Subscribe(channel string, handler MessageHandler) error { return nil }
func (r *RealRabbitMQ) Unsubscribe(channel string) error   { return nil }

func (r *RealKafka) Name() string                          { return "kafka" }
func (r *RealKafka) Connect(config *Config) error          { return nil }
func (r *RealKafka) Close() error                          { return nil }
func (r *RealKafka) Ping() error                           { return nil }
func (r *RealKafka) Publish(channel string, msg *Message) error { return nil }
func (r *RealKafka) Subscribe(channel string, handler MessageHandler) error { return nil }
func (r *RealKafka) Unsubscribe(channel string) error      { return nil }

func (r *RealNATS) Name() string                           { return "nats" }
func (r *RealNATS) Connect(config *Config) error           { return nil }
func (r *RealNATS) Close() error                           { return nil }
func (r *RealNATS) Ping() error                            { return nil }
func (r *RealNATS) Publish(channel string, msg *Message) error { return nil }
func (r *RealNATS) Subscribe(channel string, handler MessageHandler) error { return nil }
func (r *RealNATS) Unsubscribe(channel string) error       { return nil }

func NewRealRedis() *RealRedis     { return &RealRedis{} }
func NewRealRabbitMQ() *RealRabbitMQ { return &RealRabbitMQ{} }
func NewRealKafka() *RealKafka     { return &RealKafka{} }
func NewRealNATS() *RealNATS       { return &RealNATS{} }
