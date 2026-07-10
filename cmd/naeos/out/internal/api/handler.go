package api

// Handler is a small starter implementation for the api module.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}
