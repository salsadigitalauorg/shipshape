package gen

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func FactPlugin(plugins []string, pkg string, envResolver bool) {
	log.Printf("Generating fact plugin funcs for package %s: %s\n", pkg, strings.Join(plugins, ","))

	tmplPath := filepath.Join("..", "..", "pkg", "fact", "templates", "factplugin.go.tmpl")

	for _, p := range plugins {
		pluginFile := strings.ToLower(p) + "_gen.go"
		pluginDir := filepath.Join(getScriptPath(), "..", "..", "pkg", "fact")
		if pkg != "fact" {
			pluginDir = filepath.Join(pluginDir, pkg)
		}
		pluginFullFilePath := filepath.Join(pluginDir, pluginFile)
		if err := os.Remove(pluginFullFilePath); err != nil && !os.IsNotExist(err) {
			log.Fatalln(err)
		}

		templateToFile(tmplPath, struct {
			Package     string
			Plugin      string
			EnvResolver bool
		}{Package: pkg, Plugin: p, EnvResolver: envResolver}, pluginFullFilePath)
	}
}

// FactRegistry adds the Facters for a package to the registry.
func FactRegistry(pkg string) {
	log.Println("Updating Fact plugins registry - adding", pkg)

	pkgFullName := fmt.Sprintf("github.com/salsadigitalauorg/shipshape/pkg/fact/%s", pkg)

	fileLines := getFileLines(registryFullFilePath)
	if stringSliceMatch(fileLines, pkgFullName) {
		return
	}

	appendFileContent(registryFullFilePath, fmt.Sprintf("import _ \"%s\"\n", pkgFullName))
}
