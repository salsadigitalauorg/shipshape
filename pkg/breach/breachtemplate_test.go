package breach_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/salsadigitalauorg/shipshape/pkg/breach"
)

type templaterStub struct {
	template BreachTemplate
	breaches []Breach
}

func (s *templaterStub) AddBreach(b Breach) {
	s.breaches = append(s.breaches, b)
}

func (s *templaterStub) GetBreachTemplate() BreachTemplate {
	return s.template
}

// A template that references a key-value field must render cleanly against a
// *ValueBreach (which has no Key) rather than erroring the whole breach.
func TestEvaluateTemplateValueBreachWithKeyValueTemplate(t *testing.T) {
	assert := assert.New(t)

	bt := &templaterStub{
		template: BreachTemplate{
			Type:       BreachTypeKeyValue,
			Key:        "{{ .Breach.Key }}",
			ValueLabel: "Handler {{ .Breach.ValueLabel }}",
			Value:      "{{ .Breach.Value }}",
		},
	}

	EvaluateTemplate(bt, &ValueBreach{
		CheckName: "tokenised",
		Value:     "[current-user:mail]",
	}, nil)

	assert.Len(bt.breaches, 1)
	b := bt.breaches[0]
	assert.NotContains(b.String(), "unable to render")
	assert.Equal("[current-user:mail]", BreachGetValue(b))
	assert.Equal("", BreachGetKey(b))
}

// A KeyValueBreach rendered through a key-value template resolves all fields.
func TestEvaluateTemplateKeyValueBreach(t *testing.T) {
	assert := assert.New(t)

	bt := &templaterStub{
		template: BreachTemplate{
			Type:       BreachTypeKeyValue,
			Key:        "{{ .Breach.Key }}",
			ValueLabel: "Handler \"{{ .Breach.ValueLabel }}\"",
			Value:      "{{ .Breach.Value }}",
		},
	}

	EvaluateTemplate(bt, &KeyValueBreach{
		CheckName:  "tokenised",
		Key:        "contact",
		ValueLabel: "email",
		Value:      "[current-user:mail]",
	}, nil)

	assert.Len(bt.breaches, 1)
	b := bt.breaches[0]
	assert.NotContains(b.String(), "unable to render")
	assert.Equal("contact", BreachGetKey(b))
	assert.Equal("Handler \"email\"", BreachGetValueLabel(b))
	assert.Equal("[current-user:mail]", BreachGetValue(b))
}

// A KeyValuesBreach must not panic when rendered through a template; the
// switch in EvaluateTemplate previously handled only value/key-value types.
func TestEvaluateTemplateKeyValuesBreachDoesNotPanic(t *testing.T) {
	assert := assert.New(t)

	bt := &templaterStub{
		template: BreachTemplate{
			Type: BreachTypeKeyValue,
			Key:  "{{ .Breach.Key }}",
		},
	}

	kvb := &KeyValuesBreach{
		CheckName: "multi",
		Key:       "group",
		Values:    []string{"a", "b"},
	}
	assert.NotPanics(func() {
		EvaluateTemplate(bt, kvb, nil)
	})
	assert.Len(bt.breaches, 1)
	// An unhandled breach type falls back to the original breach unchanged.
	assert.Same(kvb, bt.breaches[0])
}

// A malformed template string must fail closed: EvaluateTemplateString returns
// an empty rendered field and records exactly one "unable to parse" breach
// rather than falling through to Execute on a nil template.
func TestEvaluateTemplateStringParseFailure(t *testing.T) {
	assert := assert.New(t)

	bt := &templaterStub{}
	// Unterminated action -> template.Parse fails.
	out := EvaluateTemplateString(bt, "{{ .Breach.Key ", &ValueBreach{})

	assert.Equal("", out)
	assert.Len(bt.breaches, 1)
	assert.Contains(bt.breaches[0].String(), "unable to parse breach template")
}

// No template set falls through to the raw breach unchanged.
func TestEvaluateTemplateNoTemplate(t *testing.T) {
	assert := assert.New(t)

	bt := &templaterStub{}
	raw := &ValueBreach{CheckName: "raw", Value: "v"}

	EvaluateTemplate(bt, raw, nil)

	assert.Len(bt.breaches, 1)
	assert.Same(raw, bt.breaches[0])
}
