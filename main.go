package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	_ "github.com/salsadigitalauorg/shipshape/pkg/drupal"
	_ "github.com/salsadigitalauorg/shipshape/pkg/phpstan"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"

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
	checksFiles        []string
	checkTypesToRun    []string
	excludeDb          bool
	outputFormat       string
)

func main() {
	parseFlags()
	// Parse env vars, overriding flags.
	parseEnvVars()

	if displayVersion {
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		os.Exit(0)
	}

	parseArgs()
	if !isValidOutputFormat(&outputFormat) {
		log.Fatalf("Invalid output format; needs to be one of: %s.", strings.Join(shipshape.OutputFormats, "|"))
	}

	for _, f := range checksFiles {
		if !utils.StringIsUrl(f) {
			if _, err := os.Stat(f); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "checks file '%s' not found\n", f)

				if errorCodeOnFailure {
					os.Exit(1)
				}
				os.Exit(0)
			}
		}
	}

	cfg, err := shipshape.ReadAndParseConfig(projectDir, checksFiles)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Init()
	cfg.FilterChecksToRun(checkTypesToRun, excludeDb)
	rl := cfg.RunChecks()

	switch outputFormat {
	case "json":
		data, err := json.Marshal(rl)
		if err != nil {
			log.Fatalf("Unable to convert result to json: %+v\n", err)
		}
		fmt.Println(string(data))
	case "junit":
		w := bufio.NewWriter(os.Stdout)
		rl.JUnit(w)
	case "table":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		rl.TableDisplay(w)
	case "simple":
		w := bufio.NewWriter(os.Stdout)
		rl.SimpleDisplay(w)
	}

	if rl.Status() == shipshape.Fail && errorCodeOnFailure &&
		len(rl.GetBreachesBySeverity(cfg.FailSeverity)) > 0 {

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

	pflag.BoolVarP(&errorCodeOnFailure, "error-code", "e", false, "Exit with error code if a failure is detected (env: SHIPSHAPE_ERROR_ON_FAILURE)")
	pflag.StringSliceVarP(&checksFiles, "file", "f", []string{"shipshape.yml"}, "Path to the file containing the checks. Can be specified as comma-separated single argument or using --types multiple times")
	pflag.StringVarP(&outputFormat, "output", "o", "simple", "Output format [json|junit|simple|table] (env: SHIPSHAPE_OUTPUT_FORMAT)")
	pflag.StringSliceVarP(&checkTypesToRun, "types", "t", []string(nil), "List of checks to run; default is empty, which will run all checks. Can be specified as comma-separated single argument or using --types multiple times")
	pflag.BoolVarP(&excludeDb, "exclude-db", "d", false, "Exclude checks requiring a database; overrides any db checks specified by '--types'")
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

// parseEnvVars reads and applies supported environment variables.
func parseEnvVars() {
	errorCodeOnFailureEnv := os.Getenv("SHIPSHAPE_ERROR_ON_FAILURE")
	if errorCodeOnFailureEnv != "" {
		if errorCodeOnFailureEnvBool, err := strconv.ParseBool(errorCodeOnFailureEnv); err == nil {
			errorCodeOnFailure = errorCodeOnFailureEnvBool
		}
	}

	outputFormatEnv := os.Getenv("SHIPSHAPE_OUTPUT_FORMAT")
	if outputFormatEnv != "" {
		outputFormat = outputFormatEnv
	}
}

func parseArgs() {
	args := pflag.Args()
	if len(args) > 1 {
		log.Fatalf("Max 1 argument expected, got '%+v'\n", args)
	} else if len(args) == 1 {
		projectDir = args[0]
	}
}

func isValidOutputFormat(of *string) bool {
	valid := false
	for _, fm := range shipshape.OutputFormats {
		if *of == fm {
			valid = true
			break
		}
	}
	return valid
}
