package module

type Module struct {
	Name         string            `json:"name"`
	Path         string            `json:"path,omitempty"`
	Description  string            `json:"description,omitempty"`
	Condition    string            `json:"condition,omitempty"`
	Packages     []string          `json:"packages,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}
