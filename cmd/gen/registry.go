package gen

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var registryFile = "registry_gen.go"
var registryFullFilePath = filepath.Join(getScriptPath(), "..", "..", registryFile)

func RegistryCreateFile() {
	log.Println("Cleaning up")
	os.Remove("registry_gen.go")
	createFileWithString(registryFullFilePath, "package main\n\n")
	appendFileContent(registryFullFilePath,
		"// Code generated by registry create; DO NOT EDIT.\n\n")
}

// CheckRegistry adds the checks for a package to the registry.
func CheckRegistry(pkg string) {
	log.Println("Generating checks registry - adding", pkg)

	pkgFullName := fmt.Sprintf("github.com/salsadigitalauorg/shipshape/pkg/checks/%s", pkg)

	fileLines := getFileLines(registryFullFilePath)
	if stringSliceMatch(fileLines, pkgFullName) {
		return
	}

	appendFileContent(registryFullFilePath, fmt.Sprintf("import _ \"%s\"\n", pkgFullName))
}
