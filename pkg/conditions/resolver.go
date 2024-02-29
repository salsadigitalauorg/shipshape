package conditions

import "fmt"

type Condition interface {
	FailMessage(any) string
	Satisfied(any) bool
}

type Resolver struct {
	Conditions []Condition
}

func (r *Resolver) AddCondition(c Condition) {
	r.Conditions = append(r.Conditions, c)
}

func (r *Resolver) Resolve(actual any) (bool, error) {
	for _, c := range r.Conditions {
		if !c.Satisfied(actual) {
			return false, fmt.Errorf(c.FailMessage(actual))
		}
	}
	return true, nil
}
