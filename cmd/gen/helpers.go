package gen

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func getScriptPath() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(b)
}

func createFile(fullpath string, firstTimeContent string) {
	if f, err := os.Stat(fullpath); err == nil && !f.IsDir() {
		return
	} else if !os.IsNotExist(err) {
		log.Fatalln(err)
	}

	f, err := os.OpenFile(fullpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		f.Close()
	}()

	if firstTimeContent == "" {
		return
	}

	if _, err := f.Write([]byte(firstTimeContent)); err != nil {
		log.Fatal(err)
	}
}

func getFileLines(fullpath string) []string {
	input, err := os.ReadFile(fullpath)
	if err != nil {
		log.Fatalln(err)
	}
	return strings.Split(string(input), "\n")
}

func writeFileContent(fullpath string, content string) {
	err := os.WriteFile(fullpath, []byte(content), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func appendFileContent(fullpath string, content string) {
	input, err := os.ReadFile(fullpath)
	if err != nil {
		log.Fatalln(err)
	}
	output := string(input) + content
	writeFileContent(fullpath, output)
}

func stringSliceMatch(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}
