package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

// Usage:
//
//	go run ./tools/brewgen \
//	  --version v1.0.0-alpha.1.6 \
//	  --checksums checksums.txt \
//	  --out ./tap/Formula
func main() {
	log.SetFlags(0)
	log.SetPrefix("brewgen: ")

	var (
		tag       string
		checksums string
		out       string
		repo      string
		name      string
		branch    string
		desc      string
		license   string
	)

	flag.StringVar(&tag, "version", "", "Release tag, e.g. v1.0.0-alpha.1.6 (required)")
	flag.StringVar(&checksums, "checksums", "", "Path to the release checksums.txt (required)")
	flag.StringVar(&out, "out", "", "Directory to write the formulae into (required)")
	flag.StringVar(&repo, "repo", "salsadigitalauorg/shipshape", "GitHub repository as owner/name")
	flag.StringVar(&name, "name", "shipshape", "Formula and binary name")
	flag.StringVar(&branch, "branch", "1.x", "Branch a --HEAD build compiles from")
	flag.StringVar(&desc, "desc", "Golang CLI tool to run checks/audits on your systems",
		"Formula description")
	flag.StringVar(&license, "license", "MIT", "SPDX licence identifier")
	flag.Parse()

	required := []struct {
		flag  string
		value string
	}{
		{"--version", tag},
		{"--checksums", checksums},
		{"--out", out},
	}
	for _, r := range required {
		if r.value == "" {
			log.Fatalf("%s is required", r.flag)
		}
	}

	f, err := os.Open(checksums)
	if err != nil {
		log.Fatalf("opening checksums: %v", err)
	}
	defer f.Close()

	sums, err := ParseChecksums(bufio.NewScanner(f))
	if err != nil {
		log.Fatal(err)
	}

	files, err := Render(Options{
		Tag:     tag,
		Repo:    repo,
		Name:    name,
		Branch:  branch,
		Desc:    desc,
		License: license,
	}, sums)
	if err != nil {
		log.Fatal(err)
	}

	if err := Write(out, files); err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Printf("wrote %s/%s\n", out, file.Filename)
	}
}
