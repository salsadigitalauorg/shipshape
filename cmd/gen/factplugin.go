package gen

import (
	"fmt"
	"log"
)

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
