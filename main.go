package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/lagoon"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

//go:generate go run cmd/gen.go registry create

// Version information.
var (
	version string
	commit  string
)

var (
	displayUsage   bool
	displayVersion bool
	dumpConfig     bool
	listChecks     bool
	// selfUpdate     bool

	errorCodeOnFailure bool
	projectDir         string
	checksFiles        []string
	checkTypesToRun    []string
	excludeDb          bool
	outputFormat       string
	remediate          bool
	logLevel           string
	verbose            bool
	debug              bool
	lagoonApiBaseUrl   string
	lagoonApiToken     string
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

	if listChecks {
		fmt.Println("Type of checks available:")
		checks := []string{}
		for c := range config.ChecksRegistry {
			checks = append(checks, string(c))
		}
		sort.Strings(checks)
		for _, c := range checks {
			fmt.Println("  - " + c)
		}
		os.Exit(0)
	}

	parseArgs()
	if !isValidOutputFormat(&outputFormat) {
		log.Fatalf("Invalid output format; needs to be one of: %s.", strings.Join(shipshape.OutputFormats, "|"))
	}

	determineLogLevel()

	// simple check to ensure we have everything we need to write to the API if required.
	if lagoon.PushProblemsToInsightRemote {
		if lagoonApiBaseUrl == "" {
			log.Fatal("lagoon api base url not provided")
		}
		if lagoonApiToken == "" {
			log.Fatal("lagoon api token not provided")
		}
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

	err := shipshape.Init(
		projectDir,
		checksFiles,
		checkTypesToRun,
		excludeDb,
		remediate,
		logLevel,
		lagoonApiBaseUrl,
		lagoonApiToken)
	if err != nil {
		log.Fatal(err)
	}

	if dumpConfig {
		out, err := yaml.Marshal(shipshape.RunConfig)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", string(out))
		os.Exit(0)
	}

	if shipshape.IsV2 {
		shipshape.GatherFacts()
	} else {
		shipshape.RunChecks()
	}

	switch outputFormat {
	case "json":
		data, err := json.Marshal(shipshape.RunResultList)
		if err != nil {
			log.Fatalf("Unable to convert result to json: %+v\n", err)
		}
		fmt.Println(string(data))
	case "junit":
		w := bufio.NewWriter(os.Stdout)
		shipshape.JUnit(w)
	case "table":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		shipshape.TableDisplay(w)
	case "simple":
		w := bufio.NewWriter(os.Stdout)
		shipshape.SimpleDisplay(w)
	}

	if lagoon.PushProblemsToInsightRemote {
		w := bufio.NewWriter(os.Stdout)
		err := lagoon.ProcessResultList(w, shipshape.RunResultList)
		if err != nil {
			log.Fatal(err)
		}
	}

	if shipshape.RunResultList.Status() == result.Fail && errorCodeOnFailure &&
		len(shipshape.RunResultList.GetBreachesBySeverity(string(shipshape.RunConfig.FailSeverity))) > 0 {

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
	pflag.BoolVarP(&displayVersion, "version", "", false, "Displays the application version")
	pflag.BoolVar(&dumpConfig, "dump-config", false, "Dump the final config - useful to make sure multiple config files are being merged as expected")
	pflag.BoolVar(&listChecks, "list-checks", false, "List available checks")
	// pflag.BoolVarP(&selfUpdate, "self-update", "u", false, "Updates shipshape to the latest version")

	pflag.BoolVarP(&errorCodeOnFailure, "error-code", "e", false, "Exit with error code if a failure is detected (env: SHIPSHAPE_ERROR_ON_FAILURE)")
	pflag.StringSliceVarP(&checksFiles, "file", "f", []string{"shipshape.yml"}, "Path to the file containing the checks. Can be specified as comma-separated single argument or using --types multiple times")
	pflag.StringVarP(&outputFormat, "output", "o", "simple", "Output format [json|junit|simple|table] (env: SHIPSHAPE_OUTPUT_FORMAT)")
	pflag.StringSliceVarP(&checkTypesToRun, "types", "t", []string(nil), "List of checks to run; default is empty, which will run all checks. Can be specified as comma-separated single argument or using --types multiple times")
	pflag.StringVarP(&logLevel, "log-level", "l", "warn", "Level of logs to display")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "Display verbose output - equivalent to --log-level info")
	pflag.BoolVarP(&debug, "debug", "d", false, "Display debug information - equivalent to --log-level debug")
	pflag.BoolVarP(&excludeDb, "exclude-db", "x", false, "Exclude checks requiring a database; overrides any db checks specified by '--types'")
	pflag.BoolVarP(&remediate, "remediate", "r", false, "Run remediation for supported checks")
	pflag.StringVar(&lagoonApiBaseUrl, "lagoon-api-base-url", "", "Base url for the Lagoon API when pushing problems to API (env: LAGOON_API_BASE_URL)")
	pflag.StringVar(&lagoonApiToken, "lagoon-api-token", "", "Lagoon API token when pushing problems to API (env: LAGOON_API_TOKEN)")
	pflag.BoolVar(&lagoon.PushProblemsToInsightRemote, "lagoon-push-problems-to-insights", false, "Push audit facts to Lagoon via Insights Remote")
	pflag.StringVar(&lagoon.LagoonInsightsRemoteEndpoint, "lagoon-insights-remote-endpoint", "http://lagoon-remote-insights-remote.lagoon.svc/problems", "Insights Remote Problems endpoint")
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

	lagoonApiBaseUrlEnv := os.Getenv("LAGOON_API_BASE_URL")
	if outputFormatEnv != "" {
		lagoonApiBaseUrl = lagoonApiBaseUrlEnv
	}

	lagoonApiTokenEnv := os.Getenv("LAGOON_API_TOKEN")
	if outputFormatEnv != "" {
		lagoonApiToken = lagoonApiTokenEnv
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

func determineLogLevel() {
	if debug {
		logLevel = "debug"
		return
	}

	if verbose {
		logLevel = "info"
		return
	}
}
