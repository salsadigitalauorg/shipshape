package conditions

import "fmt"

type Empty struct{}

func (c *Empty) Satisfied(actual any) bool {
	return actual.(string) == "" ||
		len(actual.([]any)) == 0 ||
		len(actual.(map[any]any)) == 0 ||
		false
}

func (c *Empty) FailMessage(actual any) string {
	return fmt.Sprintf("expected empty value, got %v", actual)
}

type NotEmpty struct{}

func (c *NotEmpty) Satisfied(actual any) bool {
	return actual.(string) != "" ||
		len(actual.([]any)) > 0 ||
		len(actual.(map[any]any)) > 0 ||
		false
}

func (c *NotEmpty) FailMessage(actual any) string {
	return "expected non-empty value, got empty"
}
