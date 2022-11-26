package gen

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

var registryFile = "registry_gen.go"
var fullRegistryFilePath string

// Registry adds the checks for a package to the registry.
func Registry(chkPkg string, chkType string, chkStruct string) {
	fullRegistryFilePath = filepath.Join(getScriptPath(), "../../", registryFile)
	createFile()
	addImportLine(chkPkg)
	addEntry(chkPkg, chkType, chkStruct)
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

	opening_code := `
package main

import "github.com/salsadigitalauorg/shipshape/pkg/shipshape"

func init() {
}
`
	if _, err := f.Write([]byte(opening_code)); err != nil {
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
	pkgFullName := fmt.Sprintf("github.com/salsadigitalauorg/shipshape/pkg/%s", chkPkg)

	fileLines := getFileLines()
	if stringSliceMatch(fileLines, pkgFullName) {
		return
	}

	importLine := fmt.Sprintf("import \"%s\"", pkgFullName)
	newFileLines := []string{}
	for i, line := range fileLines {
		if i == 3 {
			newFileLines = append(newFileLines, importLine)
		}
		newFileLines = append(newFileLines, line)
	}
	writeFileLines(newFileLines)
}

func addEntry(chkPkg string, chkType string, chkStruct string) {
	entryLine := fmt.Sprintf(
		"\tshipshape.ChecksRegistry[%s.%s] = func() shipshape.Check { return &%s.%s{} }",
		chkPkg, chkType, chkPkg, chkStruct,
	)

	fileLines := getFileLines()
	if utils.StringSliceContains(fileLines, entryLine) {
		return
	}

	newFileLines := []string{}
	for _, line := range fileLines {
		newFileLines = append(newFileLines, line)
		if line == "func init() {" {
			newFileLines = append(newFileLines, entryLine)
			continue
		}
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
