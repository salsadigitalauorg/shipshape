package breach

import (
	"bytes"
	"text/template"

	ss_rem "github.com/salsadigitalauorg/shipshape/pkg/remediation"
)

func EvaluateTemplate(bt BreachTemplater, b Breach, remediation interface{}) {
	t := bt.GetBreachTemplate()

	// No template set, use raw breach.
	if t.Type == "" {
		r := ss_rem.RemediatorFromInterface(remediation)
		b.SetRemediator(r)
		bt.AddBreach(b)
		return
	}

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

	var breachToAdd Breach
	switch rendered.Type {
	case BreachTypeValue:
		breach := b.(*ValueBreach)
		breach.ValueLabel = rendered.ValueLabel
		breach.Value = rendered.Value
		breachToAdd = breach
	case BreachTypeKeyValue:
		breach := b.(*KeyValueBreach)
		breach.KeyLabel = rendered.KeyLabel
		breach.Key = rendered.Key
		breach.ValueLabel = rendered.ValueLabel
		breach.Value = rendered.Value
		breachToAdd = breach
	}

	r := ss_rem.RemediatorFromInterface(remediation)
	breachToAdd.SetRemediator(r)
	bt.AddBreach(breachToAdd)
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
