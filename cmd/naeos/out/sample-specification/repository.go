package sample-specification

// Repository interface describes the persistence boundary for the sample-specification module.
type Repository interface {
	List() []string
}
