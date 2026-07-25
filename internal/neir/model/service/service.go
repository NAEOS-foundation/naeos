package service

type ServiceKind string

const (
	KindHTTP   ServiceKind = "http"
	KindGRPC   ServiceKind = "grpc"
	KindWorker ServiceKind = "worker"
	KindCLI    ServiceKind = "cli"
	KindJob    ServiceKind = "job"
)

type Service struct {
	Name        string            `json:"name"`
	Kind        ServiceKind       `json:"kind,omitempty"`
	Port        int               `json:"port,omitempty"`
	Description string            `json:"description,omitempty"`
	Condition   string            `json:"condition,omitempty"`
	Endpoints   []Endpoint        `json:"endpoints,omitempty"`
	Middleware  []string          `json:"middleware,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type Endpoint struct {
	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`
	Action string `json:"action,omitempty"`
}
