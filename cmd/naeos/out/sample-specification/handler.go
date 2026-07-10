package sample-specification

// Handler is a small starter implementation for the sample-specification module.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}
