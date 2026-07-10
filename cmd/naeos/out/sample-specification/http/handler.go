package http

import "fmt"

// Handler is a starter HTTP handler for the sample-specification module.
type Handler struct{}

func (h Handler) ServeHTTP(w interface{}, r interface{}) {
	fmt.Println("handler for sample-specification")
}
