package api

// Service interface describes the application behavior for the api module.
type Service interface {
	Handle() string
}
