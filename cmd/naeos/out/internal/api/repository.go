package api

// Repository interface describes the persistence boundary for the api module.
type Repository interface {
	List() []string
}
