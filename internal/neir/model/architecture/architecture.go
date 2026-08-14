package architecture

type Pattern string

const (
	PatternLayered       Pattern = "layered"
	PatternClean         Pattern = "clean"
	PatternHexagonal     Pattern = "hexagonal"
	PatternMicrokernel   Pattern = "microkernel"
	PatternEventDriven   Pattern = "event-driven"
	PatternCQRS          Pattern = "cqrs"
	PatternMonolith      Pattern = "monolith"
	PatternMonolithic    Pattern = "monolithic"
	PatternMicroservices Pattern = "microservices"
	PatternServerless    Pattern = "serverless"
)

type Architecture struct {
	Pattern     Pattern           `json:"pattern,omitempty"`
	Style       string            `json:"style,omitempty"`
	Description string            `json:"description,omitempty"`
	Principles  []string          `json:"principles,omitempty"`
	Layers      []Layer           `json:"layers,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type Layer struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Modules     []string `json:"modules,omitempty"`
}
