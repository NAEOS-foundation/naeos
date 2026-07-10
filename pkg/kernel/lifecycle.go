package kernel

type Lifecycle interface {
	Initialize() error
	Start() error
	Stop() error
}
