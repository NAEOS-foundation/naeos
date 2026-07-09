package policy

type Evaluator interface {
	Evaluate(ctx map[string]any) error
}

type DefaultEvaluator struct{}

func NewEvaluator() Evaluator {
	return DefaultEvaluator{}
}

func (DefaultEvaluator) Evaluate(ctx map[string]any) error {
	return nil
}
