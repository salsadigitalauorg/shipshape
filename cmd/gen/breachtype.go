package gen

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func BreachType(breachTypes []string) {
	log.Println("Generating breach type funcs -", strings.Join(breachTypes, ","))

	breachTypeFile := "breach_gen.go"
	breachTypeFullFilePath := filepath.Join(getScriptPath(), "..", "..", "pkg", "result", breachTypeFile)
	if err := os.Remove(breachTypeFullFilePath); err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}
	createFileWithString(breachTypeFullFilePath, "package result\n")

	for _, bt := range breachTypes {
		appendFileContent(breachTypeFullFilePath, breachTypeFuncs(bt))
	}
}

func breachTypeFuncs(bt string) string {
	tmplStr := `
/*
 * {{.BreachType}}Breach
 */
func (b *{{.BreachType}}Breach) GetCheckName() string {
	return b.CheckName
}

func (b *{{.BreachType}}Breach) GetCheckType() string {
	return b.CheckType
}

func (b *{{.BreachType}}Breach) GetRemediation() *Remediation {
	return &b.Remediation
}

func (b *{{.BreachType}}Breach) GetSeverity() string {
	return b.Severity
}

func (b *{{.BreachType}}Breach) GetType() BreachType {
	return BreachType{{.BreachType}}
}

func (b *{{.BreachType}}Breach) SetCommonValues(checkType string, checkName string, severity string) {
	b.BreachType = b.GetType()
	b.CheckType = checkType
	b.CheckName = checkName
	b.Severity = severity
}

func (b *{{.BreachType}}Breach) SetRemediation(status RemediationStatus, msg string) {
	b.Remediation.Status = status
	if msg != "" {
		b.Remediation.Messages = []string{msg}
	}
}
`
	tmpl, err := template.New("breachTypeFuncs").Parse(tmplStr)
	if err != nil {
		log.Fatalln(err)
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, struct{ BreachType string }{bt})
	if err != nil {
		log.Fatalln(err)
	}
	return buf.String()
}
