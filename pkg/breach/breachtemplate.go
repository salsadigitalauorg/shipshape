package breach

import (
	"bytes"
	"fmt"
	"text/template"
)

func EvaluateTemplate(bt BreachTemplater, b Breach) {

	t := bt.GetBreachTemplate()
	rendered := BreachTemplate{
		Type:       b.GetType(),
		ValueLabel: BreachGetValueLabel(b),
		Value:      BreachGetValue(b),
		KeyLabel:   BreachGetKeyLabel(b),
		Key:        BreachGetKey(b),
	}

	if t.KeyLabel != "" {
		rendered.KeyLabel = EvaluateTemplateString(bt, t.KeyLabel, b)
	}
	if t.Key != "" {
		rendered.Key = EvaluateTemplateString(bt, t.Key, b)
	}
	if t.ValueLabel != "" {
		rendered.ValueLabel = EvaluateTemplateString(bt, t.ValueLabel, b)
	}
	if t.Value != "" {
		rendered.Value = EvaluateTemplateString(bt, t.Value, b)
	}

	switch t.Type {
	case BreachTypeValue:
		breach, ok := b.(*ValueBreach)
		if !ok {
			bt.AddBreach(&ValueBreach{
				ValueLabel: "unable to cast breach to value",
				Value:      fmt.Sprintf("%#v", b),
			})
		}
		breach.ValueLabel = rendered.ValueLabel
		breach.Value = rendered.Value
		bt.AddBreach(breach)
	case BreachTypeKeyValue:
		breach, ok := b.(*KeyValueBreach)
		if !ok {
			bt.AddBreach(&ValueBreach{
				ValueLabel: "unable to cast breach to key-value",
				Value:      fmt.Sprintf("%#v", b),
			})
		}
		breach.KeyLabel = rendered.KeyLabel
		breach.Key = rendered.Key
		breach.ValueLabel = rendered.ValueLabel
		breach.Value = rendered.Value
		bt.AddBreach(breach)
	}
}

var TemplateFuncs = template.FuncMap{}

func EvaluateTemplateString(bt BreachTemplater, t string, b Breach) string {
	templ, err := template.New("breachTemplateString").
		Funcs(TemplateFuncs).Parse(t)
	if err != nil {
		bt.AddBreach(&ValueBreach{
			ValueLabel: "unable to parse breach template",
			Value:      err.Error(),
		})
	}

	buf := &bytes.Buffer{}
	data := struct{ Breach }{b}
	err = templ.Execute(buf, data)
	if err != nil {
		bt.AddBreach(&ValueBreach{
			ValueLabel: "unable to render breach template",
			Value:      err.Error(),
		})
	}
	return buf.String()
}
