package gen

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var registryFile = "registry_gen.go"
var fullRegistryFilePath string

// Registry adds the checks for a package to the registry.
func Registry(chkPkg string) {
	fullRegistryFilePath = filepath.Join(getScriptPath(), "../../", registryFile)
	createFile()
	addImportLine(chkPkg)
}

func getScriptPath() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(b)
}

func createFile() {
	if f, err := os.Stat(fullRegistryFilePath); err == nil && !f.IsDir() {
		return
	} else if !os.IsNotExist(err) {
		log.Fatal(err)
	}

	f, err := os.OpenFile(fullRegistryFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		f.Close()
	}()

	if _, err := f.Write([]byte("package main\n\n")); err != nil {
		log.Fatal(err)
	}
}

func getFileLines() []string {
	input, err := os.ReadFile(fullRegistryFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	return strings.Split(string(input), "\n")
}

func writeFileLines(lines []string) {
	output := strings.Join(lines, "\n")
	err := os.WriteFile(fullRegistryFilePath, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func addImportLine(chkPkg string) {
	pkgFullName := fmt.Sprintf("github.com/salsadigitalauorg/shipshape/pkg/checks/%s", chkPkg)

	fileLines := getFileLines()
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
	writeFileLines(newFileLines)
}

func stringSliceMatch(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}
