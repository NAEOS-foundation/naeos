package deployment

type Strategy string

const (
	StrategyRolling   Strategy = "rolling"
	StrategyBlueGreen Strategy = "blue-green"
	StrategyCanary    Strategy = "canary"
	StrategyRecreate  Strategy = "recreate"
)

type Deployment struct {
	Target       string            `json:"target,omitempty"`
	Strategy     Strategy          `json:"strategy,omitempty"`
	Environments []Environment     `json:"environments,omitempty"`
	Scaling      *Scaling          `json:"scaling,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

type Environment struct {
	Name      string            `json:"name"`
	Kind      string            `json:"kind,omitempty"`
	Config    map[string]string `json:"config,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
}

type Scaling struct {
	Min      int `json:"min,omitempty"`
	Max      int `json:"max,omitempty"`
	Replicas int `json:"replicas,omitempty"`
}
