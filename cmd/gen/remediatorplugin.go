package gen

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func RemediatorPlugin(plugins []string, names []string) {
	log.Println("Generating remediator plugin funcs -", strings.Join(plugins, ","))

	tmplPath := filepath.Join("..", "..", "pkg", "breach", "gen_templates", "remediatorplugin.go.tmpl")

	for i, p := range plugins {
		name := names[i]
		pluginFile := strings.ToLower(p) + "_gen.go"
		pluginFullFilePath := filepath.Join(getScriptPath(), "..", "..", "pkg", "breach", pluginFile)
		if err := os.Remove(pluginFullFilePath); err != nil && !os.IsNotExist(err) {
			log.Fatalln(err)
		}

		templateToFile(tmplPath, struct {
			Plugin string
			Name   string
		}{Plugin: p, Name: name}, pluginFullFilePath)
	}
}
