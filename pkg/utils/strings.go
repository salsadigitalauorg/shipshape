package utils

import (
	"bytes"
	"log"
	"strings"
	"text/template"
)

func MultilineOutputToSlice(output []byte) []string {
	slc := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		slc = append(slc, string(line))
	}
	return slc
}

func TemplateString(s string, data map[string]string) (string, error) {
	tmpl, err := template.New("").Parse(s)
	if err != nil {
		log.Fatalln(err)
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, data)
	if err != nil {
		log.Fatalln(err)
	}
	return buf.String(), nil
}
