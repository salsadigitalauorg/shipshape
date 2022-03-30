package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	_ "github.com/salsadigitalauorg/shipshape/pkg/drupal"
	_ "github.com/salsadigitalauorg/shipshape/pkg/phpstan"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"

	"github.com/spf13/pflag"
)

// Version information.
var (
	version string
	commit  string
)

var (
	displayUsage   bool
	displayVersion bool
	// selfUpdate     bool

	errorCodeOnFailure bool
	projectDir         string
	checksFile         string
	checkTypesToRun    []string
	outputFormat       string
)

func main() {
	parseFlags()

	if displayVersion {
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		os.Exit(0)
	}

	parseArgs()
	validateOutputFormat(&outputFormat)

	if _, err := os.Stat(checksFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "checks file '%s' not found\n", checksFile)

		if errorCodeOnFailure {
			os.Exit(1)
		}
		os.Exit(0)
	}

	cfg, err := shipshape.ReadAndParseConfig(projectDir, checksFile)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Init()
	r := cfg.RunChecks(checkTypesToRun)

	switch outputFormat {
	case "json":
		data, err := json.Marshal(r)
		if err != nil {
			log.Fatalf("Unable to convert result to json: %+v\n", err)
		}
		fmt.Println(string(data))
	case "junit":
		w := bufio.NewWriter(os.Stdout)
		r.JUnit(w)
	case "table":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		r.TableDisplay(w)
	case "simple":
		w := bufio.NewWriter(os.Stdout)
		r.SimpleDisplay(w)
	}

	if r.Status() == shipshape.Fail && errorCodeOnFailure {
		os.Exit(2)
	}
}

func parseFlags() {
	pflag.ErrHelp = errors.New("shipshape: help requested")

	pflag.Usage = func() {
		fmt.Fprint(os.Stderr, "Shipshape\n\nRun checks quickly on your project.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [dir]\n\nFlags:\n", os.Args[0])
		pflag.PrintDefaults()
	}

	pflag.BoolVarP(&displayUsage, "help", "h", false, "Displays usage information")
	pflag.BoolVarP(&displayVersion, "version", "v", false, "Displays the application version")
	// pflag.BoolVarP(&selfUpdate, "self-update", "u", false, "Updates shipshape to the latest version")

	pflag.BoolVarP(&errorCodeOnFailure, "error-code", "e", false, "Exit with error code (1) if a failure is detected")
	pflag.StringVarP(&checksFile, "file", "f", "shipshape.yml", "Path to the file containing the checks")
	pflag.StringVarP(&outputFormat, "output", "o", "simple", "Output format (json|junit|simple|table)")
	pflag.StringSliceVarP(&checkTypesToRun, "types", "t", []string(nil), "Comma-separated list of checks to run; default is empty, which will run all checks")
	pflag.Parse()

	if displayUsage {
		pflag.Usage()
		os.Exit(0)
	}

	// if selfUpdate {
	// 	internal.SelfUpdate("")
	// 	os.Exit(0)
	// }

}

func parseArgs() {
	args := pflag.Args()
	if len(args) > 1 {
		log.Fatalf("Max 1 argument expected, got '%+v'\n", args)
	} else if len(args) == 1 {
		projectDir = args[0]
	}
}

func validateOutputFormat(of *string) {
	validFormats := []string{"json", "junit", "simple", "table"}
	valid := false
	for _, fm := range validFormats {
		if *of == fm {
			valid = true
			break
		}
	}
	if !valid {
		log.Fatalf("Invalid output format; needs to be one of: %s.", strings.Join(validFormats, "|"))
	}
}
