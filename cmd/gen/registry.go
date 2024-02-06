package gen

import (
	"fmt"
	"log"
	"path/filepath"
)

// Registry adds the checks for a package to the registry.
func Registry(chkPkg string) {
	log.Println("Generating checks registry - adding", chkPkg)

	registryFile := "registry_gen.go"
	registryFullFilePath := filepath.Join(getScriptPath(), "..", "..", registryFile)
	createFile(registryFullFilePath, "package main\n\n")

	pkgFullName := fmt.Sprintf("github.com/salsadigitalauorg/shipshape/pkg/checks/%s", chkPkg)

	fileLines := getFileLines(registryFullFilePath)
	if stringSliceMatch(fileLines, pkgFullName) {
		return
	}

	importLine := fmt.Sprintf("import _ \"%s\"", pkgFullName)
	newFileLines := []string{}
	for i, line := range fileLines {
		if i == 2 {
			newFileLines = append(newFileLines, importLine)
		}
		newFileLines = append(newFileLines, line)
	}
	writeFileLines(registryFullFilePath, newFileLines)
}
