package registry

type Registry struct {
	Entries map[string]any
}

func NewRegistry() *Registry {
	return &Registry{Entries: map[string]any{}}
}
