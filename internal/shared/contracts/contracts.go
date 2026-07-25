package contracts

import (
	"errors"
	"fmt"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type Contract interface {
	Validate() error
}

type SchemaAware interface {
	SchemaVersion() string
}

type Versioned interface {
	Version() string
}

type Identifiable interface {
	ID() string
}

type Named interface {
	Name() string
}

type Describable interface {
	Description() string
}

type Observable interface {
	On(event string, handler func(args ...any))
	Emit(event string, args ...any)
}

type Cacheable interface {
	CacheKey() string
	CacheTTL() int
}

type Comparable interface {
	CompareTo(other any) int
}

type Disposable interface {
	Dispose() error
	Disposed() bool
}

type Cloneable interface {
	Clone() any
}

type Serializable interface {
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
}

type Filterable interface {
	Matches(filter Filter) bool
}

type Filter struct {
	Name       string
	Pattern    string
	Tags       []string
	MinVersion string
	MaxVersion string
}

type Sortable interface {
	LessThan(other Sortable) bool
}

type Paginated interface {
	Page() int
	PageSize() int
	TotalItems() int
	TotalPages() int
}

type PaginatedResult struct {
	Items      any
	Page       int
	PageSize   int
	TotalItems int
	TotalPages int
}

func (p PaginatedResult) HasNext() bool {
	return p.Page < p.TotalPages
}

func (p PaginatedResult) HasPrev() bool {
	return p.Page > 1
}

type Lifecycle interface {
	Start() error
	Stop() error
	IsRunning() bool
}

type Configurable interface {
	Configure(config map[string]any) error
	Config() map[string]any
}

type HealthChecker interface {
	HealthCheck() HealthStatus
}

type HealthStatus struct {
	Healthy bool              `json:"healthy"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

func Healthy(msg string) HealthStatus {
	return HealthStatus{Healthy: true, Message: msg}
}

func Unhealthy(msg string) HealthStatus {
	return HealthStatus{Healthy: false, Message: msg}
}

type MetricsProvider interface {
	MetricName() string
	MetricValue() float64
	MetricLabels() map[string]string
}

type Registry interface {
	Register(component any) error
	Unregister(name string) error
	Get(name string) (any, bool)
	List() []string
}

type InMemoryRegistry struct {
	components map[string]any
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{components: make(map[string]any)}
}

func (r *InMemoryRegistry) Register(component any) error {
	type named interface{ Name() string }
	n, ok := component.(named)
	if !ok {
		return errors.New("component must implement Name() string")
	}
	name := n.Name()
	if _, exists := r.components[name]; exists {
		return naeoserr.New(naeoserr.ErrConflict, fmt.Sprintf("component %q already registered", name))
	}
	r.components[name] = component
	return nil
}

func (r *InMemoryRegistry) Unregister(name string) error {
	if _, exists := r.components[name]; !exists {
		return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("component %q not found", name))
	}
	delete(r.components, name)
	return nil
}

func (r *InMemoryRegistry) Get(name string) (any, bool) {
	c, ok := r.components[name]
	return c, ok
}

func (r *InMemoryRegistry) List() []string {
	var names []string
	for name := range r.components {
		names = append(names, name)
	}
	return names
}

type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationWarning
}

type ValidationError struct {
	Field   string
	Message string
	Code    string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationWarning struct {
	Field   string
	Message string
}

func (v ValidationResult) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v ValidationResult) HasWarnings() bool {
	return len(v.Warnings) > 0
}

func (v ValidationResult) ErrorMessages() []string {
	var msgs []string
	for _, e := range v.Errors {
		msgs = append(msgs, e.Error())
	}
	return msgs
}

func (v ValidationResult) WarningMessages() []string {
	var msgs []string
	for _, w := range v.Warnings {
		msgs = append(msgs, w.Message)
	}
	return msgs
}

type Validator interface {
	ValidateValue(value any) ValidationResult
}

type StringValidator struct {
	MinLength int
	MaxLength int
	Pattern   string
	Enum      []string
}

func (sv StringValidator) ValidateValue(value any) ValidationResult {
	s, ok := value.(string)
	if !ok {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Field: "value", Message: "expected string", Code: "type_error"}},
		}
	}

	vr := ValidationResult{Valid: true}
	if sv.MaxLength > 0 && len(s) > sv.MaxLength {
		vr.Valid = false
		vr.Errors = append(vr.Errors, ValidationError{
			Field: "value", Message: fmt.Sprintf("exceeds max length %d", sv.MaxLength), Code: "max_length",
		})
	}
	if sv.MinLength > 0 && len(s) < sv.MinLength {
		vr.Valid = false
		vr.Errors = append(vr.Errors, ValidationError{
			Field: "value", Message: fmt.Sprintf("below min length %d", sv.MinLength), Code: "min_length",
		})
	}
	if len(sv.Enum) > 0 {
		found := false
		for _, v := range sv.Enum {
			if s == v {
				found = true
				break
			}
		}
		if !found {
			vr.Valid = false
			vr.Errors = append(vr.Errors, ValidationError{
				Field: "value", Message: fmt.Sprintf("not in allowed values: %s", strings.Join(sv.Enum, ", ")), Code: "enum",
			})
		}
	}
	return vr
}

type IntValidator struct {
	Min int
	Max int
}

func (iv IntValidator) ValidateValue(value any) ValidationResult {
	var n int
	switch v := value.(type) {
	case int:
		n = v
	case float64:
		n = int(v)
	default:
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Field: "value", Message: "expected number", Code: "type_error"}},
		}
	}

	vr := ValidationResult{Valid: true}
	if iv.Max > 0 && n > iv.Max {
		vr.Valid = false
		vr.Errors = append(vr.Errors, ValidationError{
			Field: "value", Message: fmt.Sprintf("exceeds max %d", iv.Max), Code: "max",
		})
	}
	if iv.Min != 0 && n < iv.Min {
		vr.Valid = false
		vr.Errors = append(vr.Errors, ValidationError{
			Field: "value", Message: fmt.Sprintf("below min %d", iv.Min), Code: "min",
		})
	}
	return vr
}

type ContractFunc func() error

func All(contractFns ...ContractFunc) ContractFunc {
	return func() error {
		for _, fn := range contractFns {
			if err := fn(); err != nil {
				return err
			}
		}
		return nil
	}
}

func Any(contractFns ...ContractFunc) ContractFunc {
	return func() error {
		var errs []string
		for _, fn := range contractFns {
			if err := fn(); err == nil {
				return nil
			} else {
				errs = append(errs, err.Error())
			}
		}
		return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("none of the contracts satisfied: %s", strings.Join(errs, "; ")))
	}
}

func Not(fn ContractFunc) ContractFunc {
	return func() error {
		if err := fn(); err == nil {
			return naeoserr.New(naeoserr.ErrValidation, "expected contract to fail but it passed")
		}
		return nil
	}
}

func RequireNonEmpty(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("%s must not be empty", name))
	}
	return nil
}

func RequireRange(name string, value, min, max int) error {
	if value < min || value > max {
		return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("%s must be between %d and %d, got %d", name, min, max, value))
	}
	return nil
}

func RequireOneOf(name, value string, allowed ...string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("%s must be one of [%s], got %q", name, strings.Join(allowed, ", "), value))
}

func RequireGreaterThan(name string, value, threshold int) error {
	if value <= threshold {
		return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("%s must be greater than %d, got %d", name, threshold, value))
	}
	return nil
}
