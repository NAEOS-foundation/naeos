package kernel

type EventBus interface {
	Publish(topic string, payload any)
	Subscribe(topic string, handler func(any)) error
}
