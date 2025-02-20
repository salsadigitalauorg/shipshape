package gen

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func AnalysePlugin(plugins []string) {
	log.Println("Generating analyse plugin funcs -", strings.Join(plugins, ","))

	tmplPath := filepath.Join("..", "..", "pkg", "analyse", "templates", "analyseplugin.go.tmpl")

	for _, p := range plugins {
		pluginFile := strings.ToLower(p) + "_gen.go"
		pluginFullFilePath := filepath.Join(getScriptPath(), "..", "..", "pkg", "analyse", pluginFile)
		if err := os.Remove(pluginFullFilePath); err != nil && !os.IsNotExist(err) {
			log.Fatalln(err)
		}

		templateToFile(tmplPath, struct{ Plugin string }{p}, pluginFullFilePath)
	}
}
