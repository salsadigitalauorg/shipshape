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
	tmplTestPath := filepath.Join("..", "..", "pkg", "analyse", "templates", "analyseplugin_test.go.tmpl")

	for _, p := range plugins {
		pluginFile := strings.ToLower(p) + "_gen.go"
		pluginFullFilePath := filepath.Join(getScriptPath(), "..", "..", "pkg", "analyse", pluginFile)
		if err := os.Remove(pluginFullFilePath); err != nil && !os.IsNotExist(err) {
			log.Fatalln(err)
		}

		templateToFile(tmplPath, struct{ Plugin string }{p}, pluginFullFilePath)

		// Test file.
		pluginTestFile := strings.ToLower(p) + "_gen_test.go"
		pluginFullTestFilePath := filepath.Join(getScriptPath(), "..", "..", "pkg", "analyse", pluginTestFile)
		if err := os.Remove(pluginTestFile); err != nil && !os.IsNotExist(err) {
			log.Fatalln(err)
		}

		templateToFile(tmplTestPath, struct{ Plugin string }{p}, pluginFullTestFilePath)
	}
}
