package review

type Reviewer interface {
	Review(input any) error
}

type DefaultReviewer struct{}

func NewReviewer() Reviewer {
	return DefaultReviewer{}
}

func (DefaultReviewer) Review(input any) error {
	return nil
}
