package conditions

import "fmt"

type Equal struct{ Value any }

func (c *Equal) Satisfied(actual any) bool {
	return c.Value == actual
}

func (c *Equal) FailMessage(actual any) string {
	return fmt.Sprintf("expected value to equal %v, got %v", c.Value, actual)
}

type NotEqual struct{ Value any }

func (c *NotEqual) Satisfied(actual any) bool {
	return c.Value != actual
}

func (c *NotEqual) FailMessage(actual any) string {
	return fmt.Sprintf("expected value (%v) to not equal %v", actual, c.Value)
}
