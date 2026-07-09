package renderers

type Renderer interface {
	Render(template string, data any) ([]byte, error)
}

type DefaultRenderer struct{}

func NewRenderer() Renderer {
	return DefaultRenderer{}
}

func (DefaultRenderer) Render(template string, data any) ([]byte, error) {
	return []byte(template), nil
}
