package gen

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func BreachType(breachTypes []string) {
	log.Println("Generating breach type funcs -", strings.Join(breachTypes, ","))

	tmplPath := filepath.Join("..", "..", "pkg", "breach", "gen_templates", "breachtype.go.tmpl")
	tmplTestPath := filepath.Join("..", "..", "pkg", "breach", "gen_templates", "breachtype_test.go.tmpl")

	breachTypeFile := "breach_gen.go"
	breachTypeFullFilePath := filepath.Join(getScriptPath(), "..", "..", "pkg", "breach", breachTypeFile)
	if err := os.Remove(breachTypeFullFilePath); err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}
	templateToFile(tmplPath, struct{ BreachTypes []string }{breachTypes}, breachTypeFullFilePath)

	// Test file.
	breachTypeTestFile := "breach_gen_test.go"
	breachTypeFullTestFilePath := filepath.Join(getScriptPath(), "..", "..", "pkg", "breach", breachTypeTestFile)
	if err := os.Remove(breachTypeFullTestFilePath); err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}
	templateToFile(tmplTestPath, struct{ BreachTypes []string }{breachTypes}, breachTypeFullTestFilePath)
}
