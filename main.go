package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"salsadigitalauorg/shipshape/pkg/drupal"
	"salsadigitalauorg/shipshape/pkg/shipshape"
	"text/tabwriter"

	"github.com/spf13/pflag"
)

var projectDir string
var checksFile string
var checkTypesToRun []string
var outputFormat string

func main() {
	parseFlags()
	parseArgs()
	validateOutputFormat(&outputFormat)

	drupal.RegisterChecks()

	cfg, err := shipshape.ReadAndParseConfig(projectDir, checksFile)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Init()
	r := cfg.RunChecks(checkTypesToRun)

	if outputFormat == "json" {
		data, err := json.Marshal(r)
		if err != nil {
			log.Fatalf("Unable to convert result to json: %+v\n", err)
		}
		fmt.Println(string(data))
	} else if outputFormat == "table" {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		r.TableDisplay(w)
	}

	if r.Status() == shipshape.Fail {
		os.Exit(1)
	}
}

func parseFlags() {
	pflag.StringVarP(&checksFile, "checks-file", "f", "shipshape.yml", "Path to the file containing the checks")
	pflag.StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json); default is table")
	pflag.StringSliceVarP(&checkTypesToRun, "check-types", "t", []string(nil), "Comma-separated list of checks to run; default is empty, which will run all checks")
	pflag.Parse()
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
	if *of != "json" && *of != "table" {
		log.Fatal("Invalid output format; needs to be 'table' or 'json'.")
	}
}
