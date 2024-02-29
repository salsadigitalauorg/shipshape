package conditions

import (
	"fmt"
	"regexp"
)

type RegexMatch struct {
	Regex string
	Skip  string
}

func (c *RegexMatch) Satisfied(actual any) bool {
	actualStr := actual.(string)
	if actualStr == c.Skip {
		return true
	}
	match, _ := regexp.MatchString(c.Regex, actualStr)
	return !match
}

func (c *RegexMatch) FailMessage(actual any) string {
	return fmt.Sprintf("expected value (%v) to not match regex %v", actual.(string), c.Regex)
}
