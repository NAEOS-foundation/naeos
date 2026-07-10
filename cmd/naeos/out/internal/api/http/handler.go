package http

import "fmt"

// Handler is a starter HTTP handler for the api module.
type Handler struct{}

func (h Handler) ServeHTTP(w interface{}, r interface{}) {
	fmt.Println("handler for api")
}
