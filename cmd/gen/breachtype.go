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
	breachTypeFullFilePath := filepath.Join(getScriptPath(), "..", "..", "pkg", "breach", breachTypeFile)
	if err := os.Remove(breachTypeFullFilePath); err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}
	createFileWithString(breachTypeFullFilePath, "package breach\n")

	for _, bt := range breachTypes {
		appendFileContent(breachTypeFullFilePath, breachTypeFuncs(bt))
	}
}

func breachTypeFuncs(bt string) string {
	tmplPath := filepath.Join("..", "..", "pkg", "breach", "gen_templates", "breachtype.go.tmpl")
	tmpl, err := template.ParseFiles(filepath.Join(getScriptPath(), tmplPath))
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
